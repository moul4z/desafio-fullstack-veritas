package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
)

// API agrupa os handlers HTTP e sua dependência de persistência.
type API struct {
	store *MemoryStore
}

func NewAPI(store *MemoryStore) *API {
	return &API{store: store}
}

// Routes registra os endpoints. Usa o roteamento por método do net/http
// (Go 1.22+), o que dispensa qualquer router de terceiros.
func (a *API) Routes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", a.health)
	mux.HandleFunc("GET /tasks", a.listTasks)
	mux.HandleFunc("POST /tasks", a.createTask)
	mux.HandleFunc("GET /tasks/{id}", a.getTask)
	mux.HandleFunc("PUT /tasks/{id}", a.updateTask)
	mux.HandleFunc("DELETE /tasks/{id}", a.deleteTask)

	return mux
}

// GET /health — usado pelo healthcheck do Docker e pelo aviso de "API offline"
// no frontend.
func (a *API) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// GET /tasks — lista o quadro inteiro, já na ordem das colunas.
func (a *API) listTasks(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.store.List())
}

// GET /tasks/{id} — detalhe de uma tarefa.
func (a *API) getTask(w http.ResponseWriter, r *http.Request) {
	task, err := a.store.Get(r.PathValue("id"))
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

// POST /tasks — cria uma tarefa. Título obrigatório; status padrão "todo".
func (a *API) createTask(w http.ResponseWriter, r *http.Request) {
	var in CreateTaskInput
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := in.Validate(); err != nil {
		writeValidationError(w, err)
		return
	}

	task, err := a.store.Create(in)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, task)
}

// PUT /tasks/{id} — edita texto e/ou move a tarefa de coluna e posição.
// Campos ausentes no corpo permanecem inalterados.
func (a *API) updateTask(w http.ResponseWriter, r *http.Request) {
	var in UpdateTaskInput
	if err := decodeJSON(r, &in); err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := in.Validate(); err != nil {
		writeValidationError(w, err)
		return
	}

	task, err := a.store.Update(r.PathValue("id"), in)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

// DELETE /tasks/{id} — remove a tarefa.
func (a *API) deleteTask(w http.ResponseWriter, r *http.Request) {
	if err := a.store.Delete(r.PathValue("id")); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---------- helpers de request/response ----------

const maxBodyBytes = 1 << 20 // 1 MB

// decodeJSON lê o corpo da requisição com limite de tamanho e recusa campos
// desconhecidos, devolvendo mensagens em português para o frontend exibir.
func decodeJSON(r *http.Request, dst any) error {
	if ct := r.Header.Get("Content-Type"); ct != "" && !strings.HasPrefix(ct, "application/json") {
		return errors.New("o corpo da requisição deve ser application/json")
	}

	decoder := json.NewDecoder(http.MaxBytesReader(nil, r.Body, maxBodyBytes))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return errors.New("o corpo da requisição está vazio")
		}
		return errors.New("não foi possível interpretar o JSON enviado")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if payload == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("erro ao escrever resposta: %v", err)
	}
}

// errorBody é o formato único de erro da API:
// {"error": {"message": "...", "fields": {"title": "..."}}}
type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func writeError(w http.ResponseWriter, status int, message string, fields map[string]string) {
	writeJSON(w, status, errorBody{Error: errorDetail{Message: message, Fields: fields}})
}

func writeValidationError(w http.ResponseWriter, err error) {
	var validation *ValidationError
	if errors.As(err, &validation) {
		writeError(w, http.StatusUnprocessableEntity, "não foi possível salvar: revise os campos destacados", validation.Fields)
		return
	}
	writeError(w, http.StatusBadRequest, err.Error(), nil)
}

// writeStoreError traduz erros do repositório em status HTTP.
func writeStoreError(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "tarefa não encontrada", nil)
		return
	}
	log.Printf("erro interno: %v", err)
	writeError(w, http.StatusInternalServerError, "erro interno ao processar a requisição", nil)
}
