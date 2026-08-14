package main

import (
	"strings"
	"time"
)

// Status representa a coluna do quadro em que a tarefa está.
type Status string

const (
	StatusTodo  Status = "todo"  // A Fazer
	StatusDoing Status = "doing" // Em Progresso
	StatusDone  Status = "done"  // Concluídas
)

// ColumnOrder define a ordem canônica das colunas do quadro.
// É usada tanto na validação quanto na serialização do arquivo JSON.
var ColumnOrder = []Status{StatusTodo, StatusDoing, StatusDone}

// IsValid informa se o status corresponde a uma das três colunas fixas.
func (s Status) IsValid() bool {
	for _, valid := range ColumnOrder {
		if s == valid {
			return true
		}
	}
	return false
}

// Limites de tamanho dos campos de texto. Mantidos como constantes para que
// backend e mensagens de erro fiquem em um único lugar.
const (
	MaxTitleLen       = 120
	MaxDescriptionLen = 1000
)

// Task é a tarefa do Kanban.
//
// Position define a ordem dentro da coluna (0 = topo). O backend é a fonte da
// verdade dessa ordenação: sempre que uma tarefa entra ou sai de uma coluna as
// posições são renumeradas sequencialmente.
type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	Position    int       `json:"position"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// CreateTaskInput é o corpo aceito em POST /tasks.
// Status é opcional e, quando omitido, a tarefa nasce em "A Fazer".
type CreateTaskInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      Status `json:"status"`
}

// UpdateTaskInput é o corpo aceito em PUT /tasks/{id}.
//
// Todos os campos são ponteiros para diferenciar "não enviado" de "enviado
// vazio". Isso permite usar o mesmo endpoint para editar o texto da tarefa e
// para movê-la de coluna (drag and drop envia apenas status e position).
type UpdateTaskInput struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *Status `json:"status"`
	Position    *int    `json:"position"`
}

// ValidationError agrupa os erros por campo para que o frontend consiga
// destacar o input problemático em vez de exibir apenas uma mensagem genérica.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	return "dados inválidos"
}

func newValidationError() *ValidationError {
	return &ValidationError{Fields: map[string]string{}}
}

func (e *ValidationError) add(field, message string) {
	e.Fields[field] = message
}

func (e *ValidationError) orNil() error {
	if len(e.Fields) == 0 {
		return nil
	}
	return e
}

// Validate normaliza e valida o payload de criação.
// O título é obrigatório; a descrição e o status são opcionais.
func (in *CreateTaskInput) Validate() error {
	in.Title = strings.TrimSpace(in.Title)
	in.Description = strings.TrimSpace(in.Description)

	errs := newValidationError()

	switch {
	case in.Title == "":
		errs.add("title", "o título é obrigatório")
	case len([]rune(in.Title)) > MaxTitleLen:
		errs.add("title", "o título deve ter no máximo 120 caracteres")
	}

	if len([]rune(in.Description)) > MaxDescriptionLen {
		errs.add("description", "a descrição deve ter no máximo 1000 caracteres")
	}

	if in.Status == "" {
		in.Status = StatusTodo
	} else if !in.Status.IsValid() {
		errs.add("status", "status inválido: use todo, doing ou done")
	}

	return errs.orNil()
}

// Validate normaliza e valida o payload de atualização.
// Campos ausentes são ignorados; campos presentes precisam ser válidos.
func (in *UpdateTaskInput) Validate() error {
	errs := newValidationError()

	if in.Title != nil {
		trimmed := strings.TrimSpace(*in.Title)
		in.Title = &trimmed
		switch {
		case trimmed == "":
			errs.add("title", "o título é obrigatório")
		case len([]rune(trimmed)) > MaxTitleLen:
			errs.add("title", "o título deve ter no máximo 120 caracteres")
		}
	}

	if in.Description != nil {
		trimmed := strings.TrimSpace(*in.Description)
		in.Description = &trimmed
		if len([]rune(trimmed)) > MaxDescriptionLen {
			errs.add("description", "a descrição deve ter no máximo 1000 caracteres")
		}
	}

	if in.Status != nil && !in.Status.IsValid() {
		errs.add("status", "status inválido: use todo, doing ou done")
	}

	if in.Position != nil && *in.Position < 0 {
		errs.add("position", "a posição não pode ser negativa")
	}

	if in.Title == nil && in.Description == nil && in.Status == nil && in.Position == nil {
		errs.add("_", "informe ao menos um campo para atualizar")
	}

	return errs.orNil()
}
