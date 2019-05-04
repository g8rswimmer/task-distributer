package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type createTaskPayload struct {
	Name    string   `json:"name"`
	Skills  []string `json:"skills"`
	Priorty string   `json:"priority"`
}

type agent struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}
type agentSkill struct {
	Agent string
	Skill string
}
type task struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Skills    []string  `json:"skills"`
	Priorty   string    `json:"priority"`
	Status    string    `json:"status"`
	StartTime time.Time `json:"start_time"`
	Agent     string    `json:"assigned_agent"`
}

var agents = map[string]agent{
	"1234": agent{
		ID:        "1234",
		FirstName: "Bighead",
		LastName:  "Burton",
	},
	"2345": agent{
		ID:        "2345",
		FirstName: "Ovaltine",
		LastName:  "Jenkins",
	},
	"3456": agent{
		ID:        "3456",
		FirstName: "Ground",
		LastName:  "Control",
	},
	"4567": agent{
		ID:        "4567",
		FirstName: "Jazz",
		LastName:  "Hands",
	},
}
var agentSkills = []agentSkill{
	{
		"1234",
		"skill1",
	},
	{
		"2345",
		"skill2",
	},
	{
		"2345",
		"skill3",
	},
	{
		"3456",
		"skill3",
	},
	{
		"4567",
		"skill1",
	},
	{
		"4567",
		"skill3",
	},
}
var tasks = map[string]task{}

var skills = map[string]struct{}{
	"skill1": {},
	"skill2": {},
	"skill3": {},
}
var priorities = map[string]struct{}{
	"low":  {},
	"high": {},
}

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
		err = validateSkills(taskPayload.Skills)
		if err != nil {
			formatError(writer, fmt.Sprintf("Invalid skill %s", err.Error()), http.StatusBadRequest)
			return
		}
		err = validatePriority(taskPayload.Priorty)
		if err != nil {
			formatError(writer, fmt.Sprintf("Invalid priority %s", err.Error()), http.StatusBadRequest)
			return
		}
		t, err := assignTask(*taskPayload)
		if err != nil {
			formatError(writer, fmt.Sprintf("%s", err.Error()), http.StatusInsufficientStorage)
			return
		}
		tasks[t.ID] = t
		success := struct {
			Success bool `json:"success"`
			Task    task `json:"task"`
		}{
			Success: true,
			Task:    t,
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
func assignTask(t createTaskPayload) (task, error) {
	skilledAgents, err := matchingAgents(t.Skills)
	if err != nil {
		return task{}, err
	}
	at := map[string][]task{}
	for _, skilledAgent := range skilledAgents {
		if ts, has := agentTasks(skilledAgent); has {
			at[skilledAgent.ID] = ts
		} else {
			return task{
				ID:        fmt.Sprintf("%d", rand.Int()),
				Name:      t.Name,
				Priorty:   t.Priorty,
				Skills:    t.Skills,
				Agent:     skilledAgent.ID,
				StartTime: time.Now(),
				Status:    "Assigned",
			}, nil
		}
	}

	if t.Priorty == "low" {
		return task{}, errors.New("unable to find an agent to assign the task")
	}
	a, err := recentAgent(at, "")
	if err != nil {
		return task{}, err
	}

	return task{
		ID:        fmt.Sprintf("%d", rand.Int()),
		Name:      t.Name,
		Priorty:   t.Priorty,
		Skills:    t.Skills,
		Agent:     a.ID,
		StartTime: time.Now(),
		Status:    "Assigned",
	}, nil
}
func recentAgent(aTasks map[string][]task, priority string) (agent, error) {
	agentID := ""
	startTime := time.Time{}
	for id, ts := range aTasks {
		if len(ts) == 1 {
			t := ts[0]
			if t.Priorty == "low" {
				if t.StartTime.After(startTime) {
					startTime = t.StartTime
					agentID = id
				}
			}
		}
	}
	if agentID == "" {
		return agent{}, errors.New("unable to find an agent to assign the task")
	}
	return agents[agentID], nil
}
func agentTasks(a agent) ([]task, bool) {
	var ts []task
	for _, t := range tasks {
		if t.Agent == a.ID {
			ts = append(ts, t)
		}
	}
	if len(ts) == 0 {
		return nil, false
	}
	return ts, true
}
func matchingAgents(skills []string) ([]agent, error) {
	var as []agent
	fmt.Printf("%+v\n", skills)
	for _, a := range agents {
		aSkills := retrieveAgentSkills(a)
		fmt.Printf("%+v\n", aSkills)
		found := true
		for _, skill := range skills {
			if _, has := aSkills[skill]; has == false {
				found = false
			}
		}
		if found {
			as = append(as, a)
		}
	}
	fmt.Printf("%+v\n", as)
	if len(as) == 0 {
		return nil, errors.New("no agents have the skills")
	}
	return as, nil
}
func retrieveAgentSkills(a agent) map[string]struct{} {
	skillSet := make(map[string]struct{})
	for _, as := range agentSkills {
		if a.ID == as.Agent {
			skillSet[as.Skill] = struct{}{}
		}
	}
	return skillSet
}
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

func createPayload(payload io.ReadCloser) (*createTaskPayload, error) {
	decoder := json.NewDecoder(payload)
	defer payload.Close()
	var task createTaskPayload
	err := decoder.Decode(&task)
	if err != nil {
		return nil, err
	}

	return &task, nil
}
func (payload *createTaskPayload) requiredFields() error {
	if payload.Name == "" {
		return errors.New("name field must be present")
	}
	if payload.Skills == nil {
		return errors.New("skills field must be present")
	}
	if payload.Priorty == "" {
		return errors.New("priority field must be present")
	}
	return nil
}
func validateSkills(taskSkills []string) error {
	for _, ts := range taskSkills {
		if _, has := skills[ts]; has == false {
			return fmt.Errorf("task skill is not supported %s", ts)
		}
	}
	return nil
}
func validatePriority(priority string) error {
	if _, has := priorities[priority]; has == false {
		return fmt.Errorf("task priority is not supported %s", priority)
	}
	return nil
}

func statusTaskHandler(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		routes := strings.Split(request.URL.String(), "/")
		taskID := routes[len(routes)-1]
		if taskID == "" {
			formatError(writer, "Task Id must be included in the URL", http.StatusBadRequest)
			return
		}
		if t, has := tasks[taskID]; has {
			success := struct {
				Success bool `json:"success"`
				Task    task `json:"task"`
			}{
				Success: true,
				Task:    t,
			}
			resp, err := json.Marshal(success)
			if err != nil {
				formatError(writer, fmt.Sprintf("Unable to encode response %s", err.Error()), http.StatusInternalServerError)
				return
			}
			writer.Header().Set("Content-Type", "application/json")
			writer.Write(resp)
		} else {
			formatError(writer, fmt.Sprintf("Task %s is not present", taskID), http.StatusBadRequest)
			return
		}
	default:
		http.Error(writer, fmt.Sprintf("Method is not supported %s", request.Method), http.StatusMethodNotAllowed)
	}
}

func completeTaskHandler(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		routes := strings.Split(request.URL.String(), "/")
		taskID := routes[len(routes)-1]
		if taskID == "" {
			formatError(writer, "Task Id must be included in the URL", http.StatusBadRequest)
			return
		}
		if t, has := tasks[taskID]; has {
			t.Status = "Complete"
			tasks[taskID] = t
			success := struct {
				Success bool `json:"success"`
				Task    task `json:"task"`
			}{
				Success: true,
				Task:    t,
			}
			resp, err := json.Marshal(success)
			if err != nil {
				formatError(writer, fmt.Sprintf("Unable to encode response %s", err.Error()), http.StatusInternalServerError)
				return
			}
			writer.Header().Set("Content-Type", "application/json")
			writer.Write(resp)
		} else {
			formatError(writer, fmt.Sprintf("Task %s is not present", taskID), http.StatusBadRequest)
			return
		}
	default:
		http.Error(writer, fmt.Sprintf("Method is not supported %s", request.Method), http.StatusMethodNotAllowed)
	}
}

func currentTasks(a agent) ([]task, bool) {
	var ts []task
	for _, t := range tasks {
		if t.Agent == a.ID && t.Status != "Complete" {
			ts = append(ts, t)
		}
	}
	if len(ts) == 0 {
		return nil, false
	}
	return ts, true
}

type listAgentTasks struct {
	agent
	Tasks []task `json:"tasks"`
}

func listAgentHandler(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		var ats []listAgentTasks
		for _, a := range agents {
			if ts, has := currentTasks(a); has {
				at := listAgentTasks{
					agent: a,
					Tasks: ts,
				}
				ats = append(ats, at)
			}
		}
		success := struct {
			Success    bool             `json:"success"`
			AgentTasks []listAgentTasks `json:"agent_tasks"`
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
