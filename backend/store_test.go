package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func mustCreate(t *testing.T, store *MemoryStore, title string, status Status) Task {
	t.Helper()
	in := CreateTaskInput{Title: title, Status: status}
	if err := in.Validate(); err != nil {
		t.Fatalf("payload de teste inválido: %v", err)
	}
	task, err := store.Create(in)
	if err != nil {
		t.Fatalf("Create(%q) retornou erro: %v", title, err)
	}
	return task
}

func titlesIn(t *testing.T, store *MemoryStore, status Status) []string {
	t.Helper()
	var out []string
	for _, task := range store.List() {
		if task.Status == status {
			out = append(out, task.Title)
		}
	}
	return out
}

func TestCreateAssignsSequentialPositions(t *testing.T) {
	store := NewMemoryStore("")

	first := mustCreate(t, store, "primeira", StatusTodo)
	second := mustCreate(t, store, "segunda", StatusTodo)

	if first.Position != 0 || second.Position != 1 {
		t.Fatalf("posições esperadas 0 e 1, obtidas %d e %d", first.Position, second.Position)
	}
	if first.ID == second.ID {
		t.Fatal("ids duplicados gerados pelo store")
	}
}

func TestCreateInputValidation(t *testing.T) {
	cases := map[string]struct {
		input     CreateTaskInput
		wantField string
	}{
		"título vazio":       {CreateTaskInput{Title: "   "}, "title"},
		"status inexistente": {CreateTaskInput{Title: "ok", Status: "arquivada"}, "status"},
		"título muito longo": {CreateTaskInput{Title: string(make([]byte, 121))}, "title"},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.input.Validate()

			var validation *ValidationError
			if !errors.As(err, &validation) {
				t.Fatalf("esperava ValidationError, obtido %v", err)
			}
			if _, ok := validation.Fields[tc.wantField]; !ok {
				t.Fatalf("esperava erro no campo %q, obtido %v", tc.wantField, validation.Fields)
			}
		})
	}
}

func TestCreateDefaultsToTodo(t *testing.T) {
	in := CreateTaskInput{Title: "sem status"}
	if err := in.Validate(); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if in.Status != StatusTodo {
		t.Fatalf("status padrão esperado %q, obtido %q", StatusTodo, in.Status)
	}
}

func TestUpdateMovesTaskBetweenColumns(t *testing.T) {
	store := NewMemoryStore("")
	task := mustCreate(t, store, "mover", StatusTodo)
	mustCreate(t, store, "fica", StatusTodo)

	doing := StatusDoing
	updated, err := store.Update(task.ID, UpdateTaskInput{Status: &doing})
	if err != nil {
		t.Fatalf("Update retornou erro: %v", err)
	}

	if updated.Status != StatusDoing || updated.Position != 0 {
		t.Fatalf("esperava doing/0, obtido %s/%d", updated.Status, updated.Position)
	}

	// A coluna de origem precisa fechar o buraco deixado pela tarefa.
	remaining := store.List()
	for _, current := range remaining {
		if current.Status == StatusTodo && current.Position != 0 {
			t.Fatalf("coluna de origem não foi renumerada: %+v", current)
		}
	}
}

func TestUpdateReordersInsideSameColumn(t *testing.T) {
	store := NewMemoryStore("")
	mustCreate(t, store, "A", StatusTodo)
	mustCreate(t, store, "B", StatusTodo)
	last := mustCreate(t, store, "C", StatusTodo)

	top := 0
	if _, err := store.Update(last.ID, UpdateTaskInput{Position: &top}); err != nil {
		t.Fatalf("Update retornou erro: %v", err)
	}

	got := titlesIn(t, store, StatusTodo)
	want := []string{"C", "A", "B"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ordem esperada %v, obtida %v", want, got)
		}
	}
}

func TestUpdateAndGetUnknownIDReturnNotFound(t *testing.T) {
	store := NewMemoryStore("")
	title := "x"

	if _, err := store.Get("inexistente"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get: esperava ErrNotFound, obtido %v", err)
	}
	if _, err := store.Update("inexistente", UpdateTaskInput{Title: &title}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Update: esperava ErrNotFound, obtido %v", err)
	}
	if err := store.Delete("inexistente"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Delete: esperava ErrNotFound, obtido %v", err)
	}
}

func TestDeleteRenumbersColumn(t *testing.T) {
	store := NewMemoryStore("")
	first := mustCreate(t, store, "A", StatusDoing)
	mustCreate(t, store, "B", StatusDoing)

	if err := store.Delete(first.ID); err != nil {
		t.Fatalf("Delete retornou erro: %v", err)
	}

	tasks := store.List()
	if len(tasks) != 1 {
		t.Fatalf("esperava 1 tarefa restante, obtidas %d", len(tasks))
	}
	if tasks[0].Position != 0 {
		t.Fatalf("posição esperada 0 após remoção, obtida %d", tasks[0].Position)
	}
}

func TestPersistenceSurvivesRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")

	original := NewMemoryStore(path)
	mustCreate(t, original, "persistida", StatusDoing)

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("arquivo de dados não foi criado: %v", err)
	}

	// Simula reiniciar o processo apontando para o mesmo arquivo.
	restarted := NewMemoryStore(path)
	if err := restarted.Load(); err != nil {
		t.Fatalf("Load retornou erro: %v", err)
	}

	tasks := restarted.List()
	if len(tasks) != 1 || tasks[0].Title != "persistida" || tasks[0].Status != StatusDoing {
		t.Fatalf("estado não foi restaurado corretamente: %+v", tasks)
	}
}

func TestLoadWithoutFileStartsEmpty(t *testing.T) {
	store := NewMemoryStore(filepath.Join(t.TempDir(), "ainda-nao-existe.json"))

	if err := store.Load(); err != nil {
		t.Fatalf("arquivo ausente não deveria ser erro: %v", err)
	}
	if !store.IsEmpty() {
		t.Fatal("quadro deveria começar vazio")
	}
}
