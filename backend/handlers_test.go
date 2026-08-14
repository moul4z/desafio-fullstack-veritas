package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestServer() http.Handler {
	store := NewMemoryStore("") // apenas memória: nada é escrito em disco nos testes
	return withCORS(NewAPI(store).Routes())
}

func do(t *testing.T, handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	var reader *bytes.Reader
	if body == "" {
		reader = bytes.NewReader(nil)
	} else {
		reader = bytes.NewReader([]byte(body))
	}

	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	return res
}

func decodeTask(t *testing.T, res *httptest.ResponseRecorder) Task {
	t.Helper()
	var task Task
	if err := json.Unmarshal(res.Body.Bytes(), &task); err != nil {
		t.Fatalf("resposta não é uma tarefa válida: %v (%s)", err, res.Body.String())
	}
	return task
}

func TestTaskCRUDLifecycle(t *testing.T) {
	handler := newTestServer()

	// CREATE
	created := do(t, handler, http.MethodPost, "/tasks", `{"title":"Nova tarefa","description":"detalhe"}`)
	if created.Code != http.StatusCreated {
		t.Fatalf("POST /tasks: esperava 201, obtido %d (%s)", created.Code, created.Body.String())
	}
	task := decodeTask(t, created)
	if task.Status != StatusTodo {
		t.Fatalf("tarefa deveria nascer em todo, obtido %q", task.Status)
	}

	// READ (lista)
	list := do(t, handler, http.MethodGet, "/tasks", "")
	var tasks []Task
	if err := json.Unmarshal(list.Body.Bytes(), &tasks); err != nil || len(tasks) != 1 {
		t.Fatalf("GET /tasks: esperava 1 tarefa, obtido %s", list.Body.String())
	}

	// UPDATE (move para "Em Progresso")
	updated := do(t, handler, http.MethodPut, "/tasks/"+task.ID, `{"status":"doing"}`)
	if updated.Code != http.StatusOK {
		t.Fatalf("PUT /tasks/{id}: esperava 200, obtido %d (%s)", updated.Code, updated.Body.String())
	}
	if moved := decodeTask(t, updated); moved.Status != StatusDoing || moved.Title != "Nova tarefa" {
		t.Fatalf("update parcial não preservou o título: %+v", moved)
	}

	// DELETE
	deleted := do(t, handler, http.MethodDelete, "/tasks/"+task.ID, "")
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("DELETE /tasks/{id}: esperava 204, obtido %d", deleted.Code)
	}

	if after := do(t, handler, http.MethodGet, "/tasks/"+task.ID, ""); after.Code != http.StatusNotFound {
		t.Fatalf("GET após remoção: esperava 404, obtido %d", after.Code)
	}
}

func TestCreateRejectsEmptyTitle(t *testing.T) {
	handler := newTestServer()

	res := do(t, handler, http.MethodPost, "/tasks", `{"title":"   "}`)
	if res.Code != http.StatusUnprocessableEntity {
		t.Fatalf("esperava 422, obtido %d (%s)", res.Code, res.Body.String())
	}

	var body errorBody
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("erro fora do formato esperado: %v", err)
	}
	if body.Error.Fields["title"] == "" {
		t.Fatalf("esperava detalhe do campo title, obtido %+v", body.Error)
	}
}

func TestCreateRejectsInvalidStatus(t *testing.T) {
	handler := newTestServer()

	res := do(t, handler, http.MethodPost, "/tasks", `{"title":"ok","status":"arquivada"}`)
	if res.Code != http.StatusUnprocessableEntity {
		t.Fatalf("esperava 422, obtido %d (%s)", res.Code, res.Body.String())
	}
}

func TestMalformedJSONReturns400(t *testing.T) {
	handler := newTestServer()

	res := do(t, handler, http.MethodPost, "/tasks", `{"title":`)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("esperava 400, obtido %d (%s)", res.Code, res.Body.String())
	}
}

func TestUnknownTaskReturns404(t *testing.T) {
	handler := newTestServer()

	res := do(t, handler, http.MethodPut, "/tasks/tsk_inexistente", `{"title":"x"}`)
	if res.Code != http.StatusNotFound {
		t.Fatalf("esperava 404, obtido %d", res.Code)
	}
}

func TestCORSPreflight(t *testing.T) {
	handler := newTestServer()

	req := httptest.NewRequest(http.MethodOptions, "/tasks", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")

	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("preflight: esperava 204, obtido %d", res.Code)
	}
	if got := res.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("cabeçalho de origem ausente ou incorreto: %q", got)
	}
	if methods := res.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(methods, "DELETE") {
		t.Fatalf("DELETE deveria estar liberado, obtido %q", methods)
	}
}
