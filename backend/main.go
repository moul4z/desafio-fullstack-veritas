package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	port := env("PORT", "8080")
	dataFile := env("DATA_FILE", "data/tasks.json")

	store := NewMemoryStore(dataFile)
	if err := store.Load(); err != nil {
		log.Printf("não foi possível carregar o arquivo de tarefas: %v", err)
	}

	if env("SEED", "false") == "true" && store.IsEmpty() {
		seedExampleTasks(store)
	}

	handler := withCORS(NewAPI(store).Routes())

	log.Printf("API disponível em http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}

func seedExampleTasks(store *MemoryStore) {
	examples := []CreateTaskInput{
		{Title: "Ler as orientações do desafio", Description: "Conferir o escopo mínimo e os critérios de avaliação.", Status: StatusDone},
		{Title: "Modelar a API de tarefas", Description: "Definir as rotas e os status válidos.", Status: StatusDone},
		{Title: "Ajustar responsividade do quadro", Description: "Conferir o comportamento em telas menores.", Status: StatusDoing},
		{Title: "Escrever o README", Description: "Documentar execução e decisões técnicas.", Status: StatusTodo},
		{Title: "Revisar o fluxo do usuário", Description: "Testar criação, edição, movimentação e exclusão.", Status: StatusTodo},
	}

	for _, input := range examples {
		if err := input.Validate(); err != nil {
			continue
		}
		if _, err := store.Create(input); err != nil {
			log.Printf("não foi possível criar tarefa de exemplo: %v", err)
			return
		}
	}
}

func env(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
