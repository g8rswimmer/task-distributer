package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	http.HandleFunc("/v1/task/create", createTaskHandler)
	http.HandleFunc("/v1/task/status/", statusTaskHandler)
	http.HandleFunc("/v1/task/complete/", completeTaskHandler)

	http.HandleFunc("/v1/agent/list", completeTaskHandler)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
