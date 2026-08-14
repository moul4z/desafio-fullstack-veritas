package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ErrNotFound é retornado quando o id informado não existe no repositório.
var ErrNotFound = errors.New("tarefa não encontrada")

// MemoryStore guarda as tarefas em memória, protegido por RWMutex, e — quando
// filePath está preenchido — espelha o estado em um arquivo JSON após cada
// escrita. Assim o requisito obrigatório (memória) e o bônus (arquivo)
// convivem sem duas implementações separadas.
type MemoryStore struct {
	mu       sync.RWMutex
	tasks    map[string]*Task
	order    map[Status][]string // ids de cada coluna, do topo para a base
	filePath string
}

// NewMemoryStore cria o repositório. Se filePath for vazio, o modo é apenas
// memória e os dados somem quando o processo encerra.
func NewMemoryStore(filePath string) *MemoryStore {
	s := &MemoryStore{
		tasks:    map[string]*Task{},
		order:    map[Status][]string{},
		filePath: filePath,
	}
	for _, status := range ColumnOrder {
		s.order[status] = []string{}
	}
	return s
}

// ---------- leitura ----------

// List devolve todas as tarefas na ordem do quadro: coluna por coluna, e
// dentro de cada coluna da primeira para a última posição.
func (s *MemoryStore) List() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.snapshot()
}

// Get devolve uma tarefa pelo id.
func (s *MemoryStore) Get(id string) (Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, ok := s.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
	}
	return *task, nil
}

// ---------- escrita ----------

// Create insere a tarefa no fim da coluna indicada.
func (s *MemoryStore) Create(in CreateTaskInput) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	task := &Task{
		ID:          newID(),
		Title:       in.Title,
		Description: in.Description,
		Status:      in.Status,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.tasks[task.ID] = task
	s.order[task.Status] = append(s.order[task.Status], task.ID)
	s.reindex(task.Status)

	if err := s.persist(); err != nil {
		return Task{}, err
	}
	return *task, nil
}

// Update aplica os campos enviados. Quando status ou position mudam, a tarefa
// é reposicionada e as duas colunas afetadas são renumeradas.
func (s *MemoryStore) Update(id string, in UpdateTaskInput) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
	}

	if in.Title != nil {
		task.Title = *in.Title
	}
	if in.Description != nil {
		task.Description = *in.Description
	}

	if in.Status != nil || in.Position != nil {
		target := task.Status
		if in.Status != nil {
			target = *in.Status
		}

		// Sem position explícita: manter o índice atual quando a coluna não
		// muda, ou ir para o fim da coluna de destino quando muda.
		index := len(s.order[target])
		if in.Position != nil {
			index = *in.Position
		} else if target == task.Status {
			index = indexOf(s.order[target], id)
		}

		// O detach encurta a coluna de origem, então o clamp precisa vir
		// depois dele para que arrastar até o fim da própria coluna funcione.
		s.detach(task.Status, id)
		s.insertAt(target, id, clamp(index, 0, len(s.order[target])))

		previous := task.Status
		task.Status = target
		s.reindex(previous)
		s.reindex(target)
	}

	task.UpdatedAt = time.Now().UTC()

	if err := s.persist(); err != nil {
		return Task{}, err
	}
	return *task, nil
}

// Delete remove a tarefa e fecha o buraco deixado na coluna.
func (s *MemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, ok := s.tasks[id]
	if !ok {
		return ErrNotFound
	}

	s.detach(task.Status, id)
	delete(s.tasks, id)
	s.reindex(task.Status)

	return s.persist()
}

// ---------- helpers internos (sempre chamados com o lock em mãos) ----------

func (s *MemoryStore) snapshot() []Task {
	out := make([]Task, 0, len(s.tasks))
	for _, status := range ColumnOrder {
		for _, id := range s.order[status] {
			if task, ok := s.tasks[id]; ok {
				out = append(out, *task)
			}
		}
	}
	return out
}

func (s *MemoryStore) detach(status Status, id string) {
	ids := s.order[status]
	for i, current := range ids {
		if current == id {
			s.order[status] = append(ids[:i], ids[i+1:]...)
			return
		}
	}
}

func (s *MemoryStore) insertAt(status Status, id string, index int) {
	ids := s.order[status]
	index = clamp(index, 0, len(ids))
	ids = append(ids, "")
	copy(ids[index+1:], ids[index:])
	ids[index] = id
	s.order[status] = ids
}

// reindex mantém Task.Position coerente com a posição real na fatia.
func (s *MemoryStore) reindex(status Status) {
	for i, id := range s.order[status] {
		if task, ok := s.tasks[id]; ok {
			task.Position = i
		}
	}
}

func indexOf(ids []string, id string) int {
	for i, current := range ids {
		if current == id {
			return i
		}
	}
	return len(ids)
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func newID() string {
	return fmt.Sprintf("tsk_%d", time.Now().UnixNano())
}

// ---------- persistência em arquivo (bônus) ----------

// persist salva um snapshot do quadro em JSON quando DATA_FILE está configurado.
func (s *MemoryStore) persist() error {
	if s.filePath == "" {
		return nil
	}

	data, err := json.MarshalIndent(s.snapshot(), "", "  ")
	if err != nil {
		return fmt.Errorf("serializar tarefas: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(s.filePath), 0o755); err != nil {
		return fmt.Errorf("criar diretório de dados: %w", err)
	}

	if err := os.WriteFile(s.filePath, data, 0o644); err != nil {
		return fmt.Errorf("gravar arquivo de dados: %w", err)
	}
	return nil
}

// Load lê o arquivo JSON, se existir, e reconstrói memória e ordenação.
// Arquivo ausente não é erro: significa apenas que o quadro começa vazio.
func (s *MemoryStore) Load() error {
	if s.filePath == "" {
		return nil
	}

	data, err := os.ReadFile(s.filePath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("ler arquivo de dados: %w", err)
	}

	var tasks []Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return fmt.Errorf("interpretar arquivo de dados: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.tasks = map[string]*Task{}
	for _, status := range ColumnOrder {
		s.order[status] = []string{}
	}

	for i := range tasks {
		task := tasks[i]
		if !task.Status.IsValid() {
			task.Status = StatusTodo // tolera arquivo editado à mão
		}
		s.tasks[task.ID] = &task
		s.order[task.Status] = append(s.order[task.Status], task.ID)
	}
	for _, status := range ColumnOrder {
		s.reindex(status)
	}
	return nil
}

// IsEmpty informa se o quadro não tem nenhuma tarefa (usado pelo seed).
func (s *MemoryStore) IsEmpty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.tasks) == 0
}
