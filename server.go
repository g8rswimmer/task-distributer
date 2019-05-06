package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

var destributerDb *sql.DB

func formatError(writer http.ResponseWriter, message string, status int) {
	errorResponse := struct {
		Success      bool   `json:"success"`
		ErrorMessage string `json:"error_message"`
	}{
		Success:      false,
		ErrorMessage: message,
	}
	resp, err := json.Marshal(errorResponse)
	if err != nil {
		http.Error(writer, fmt.Sprintf(`{"success": false, "message": %s`, err.Error()), http.StatusInternalServerError)
		return
	}
	http.Error(writer, string(resp), status)
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	var err error
	destributerDb, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("error opening database: %q", err)
	}

	http.HandleFunc("/v1/task/create", createTaskHandler)
	http.HandleFunc("/v1/task/", statusTaskHandler)
	http.HandleFunc("/v1/task/complete/", completeTaskHandler)

	http.HandleFunc("/v1/agent/list", listAgentHandler)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
