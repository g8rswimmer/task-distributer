package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// createTaskHandler will attempt to create and distribute a task to an agent.
func createTaskHandler(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodPost:
		taskPayload, err := createPayload(request.Body)
		if err != nil {
			formatError(writer, fmt.Sprintf("Unable to decode payload %s", err.Error()), http.StatusInternalServerError)
			return
		}
		err = taskPayload.requiredFields()
		if err != nil {
			formatError(writer, fmt.Sprintf("Required field missing %s", err.Error()), http.StatusBadRequest)
			return
		}
		err = taskPayload.validateSkills(destributerDb)
		if err != nil {
			formatError(writer, fmt.Sprintf("Invalid skill %s", err.Error()), http.StatusBadRequest)
			return
		}
		err = taskPayload.validatePriority(destributerDb)
		if err != nil {
			formatError(writer, fmt.Sprintf("Invalid priority %s", err.Error()), http.StatusBadRequest)
			return
		}
		t := &task{
			db: destributerDb,
		}
		err = t.assignTask(*taskPayload)
		if err != nil {
			formatError(writer, fmt.Sprintf("%s", err.Error()), http.StatusInsufficientStorage)
			return
		}
		success := struct {
			Success bool `json:"success"`
			Task    task `json:"task"`
		}{
			Success: true,
			Task:    *t,
		}
		resp, err := json.Marshal(success)
		if err != nil {
			formatError(writer, fmt.Sprintf("Unable to encode response %s", err.Error()), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.Write(resp)
	default:
		http.Error(writer, fmt.Sprintf("Method is not supported %s", request.Method), http.StatusMethodNotAllowed)
	}
}

// statusTaskHandler will return the current status of the task.
func statusTaskHandler(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		routes := strings.Split(request.URL.String(), "/")
		taskID := routes[len(routes)-1]
		if taskID == "" {
			formatError(writer, "Task Id must be included in the URL", http.StatusBadRequest)
			return
		}
		t := &task{
			db: destributerDb,
		}
		err := t.retrieve(taskID)
		if err != nil {
			formatError(writer, fmt.Sprintf("Task %s is not present", taskID), http.StatusBadRequest)
			return
		}
		success := struct {
			Success bool `json:"success"`
			Task    task `json:"task"`
		}{
			Success: true,
			Task:    *t,
		}
		resp, err := json.Marshal(success)
		if err != nil {
			formatError(writer, fmt.Sprintf("Unable to encode response %s", err.Error()), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.Write(resp)
	default:
		http.Error(writer, fmt.Sprintf("Method is not supported %s", request.Method), http.StatusMethodNotAllowed)
	}
}

// completeTaskHandler sets the task as completed.
func completeTaskHandler(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		routes := strings.Split(request.URL.String(), "/")
		taskID := routes[len(routes)-1]
		if taskID == "" {
			formatError(writer, "Task Id must be included in the URL", http.StatusBadRequest)
			return
		}
		err := updateTaskStatus(destributerDb, taskID, "Complete")
		if err != nil {
			formatError(writer, fmt.Sprintf("Task %s is not present", taskID), http.StatusBadRequest)
			return
		}

		success := struct {
			Success bool `json:"success"`
		}{
			Success: true,
		}
		resp, err := json.Marshal(success)
		if err != nil {
			formatError(writer, fmt.Sprintf("Unable to encode response %s", err.Error()), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.Write(resp)
	default:
		http.Error(writer, fmt.Sprintf("Method is not supported %s", request.Method), http.StatusMethodNotAllowed)
	}
}

// listAgentHandler will list the agents and what they are currently working on
func listAgentHandler(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		ats, err := retrieveAgentTasks(destributerDb)
		if err != nil {
			formatError(writer, fmt.Sprintf("Unable to encode response %s", err.Error()), http.StatusInternalServerError)
		}

		success := struct {
			Success    bool         `json:"success"`
			AgentTasks []agentTasks `json:"agent_tasks"`
		}{
			Success:    true,
			AgentTasks: ats,
		}
		resp, err := json.Marshal(success)
		if err != nil {
			formatError(writer, fmt.Sprintf("Unable to encode response %s", err.Error()), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.Write(resp)

	default:
		http.Error(writer, fmt.Sprintf("Method is not supported %s", request.Method), http.StatusMethodNotAllowed)
	}
}
