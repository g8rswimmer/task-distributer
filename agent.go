package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// agent is the payload for the database and HTTP response
type agent struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// agents handles the methods for multiple agents
type agents struct {
	db *sql.DB
}

// agentTasks is the list of tasks for an agent.
type agentTasks struct {
	agent
	Tasks []task `json:"tasks,omitempty"`
}

func (a *agents) retrieve(ids []string) ([]agent, error) {
	i := make([]string, len(ids))
	for idx, id := range ids {
		i[idx] = fmt.Sprintf("'%s'", id)
	}

	stmt := fmt.Sprintf(`SELECT ID, FIRSTNAME, LASTNAME FROM AGENTS WHERE ID IN (%s)`, strings.Join(i, ","))
	rows, err := a.db.Query(stmt)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer rows.Close()
	var agents []agent
	for rows.Next() {
		var a agent
		if err := rows.Scan(&a.ID, &a.FirstName, &a.LastName); err != nil {
			return nil, errors.New("no agents found")
		}
		agents = append(agents, a)
	}

	fmt.Printf("%+v\n", agents)
	if len(agents) == 0 {
		return nil, errors.New("no agents found")
	}

	return agents, nil
}

func (a *agents) tasks(agents []agent) (map[string][]task, error) {
	ids := make([]string, len(agents))
	for idx, a := range agents {
		ids[idx] = fmt.Sprintf("'%s'", a.ID)
	}

	stmt := `
	SELECT
	Id, Createdate, name, PRIORITIES.priority_level, agent
	FROM tasks
	INNER JOIN PRIORITIES ON tasks.priority = PRIORITIES.priority
	WHERE 
		agent IN (%s)
	AND
		status = 'Assigned'	 	
	`
	formattedStmt := fmt.Sprintf(stmt, strings.Join(ids, ","))
	fmt.Println(formattedStmt)
	rows, err := a.db.Query(formattedStmt)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer rows.Close()
	at := map[string][]task{}
	for rows.Next() {
		var t task
		if err := rows.Scan(&t.ID, &t.StartTime, &t.Name, &t.priorityLevel, &t.Agent); err != nil {
			return nil, errors.New("no agents found")
		}
		var ts []task
		if tks, has := at[t.Agent]; has {
			ts = tks
		}
		ts = append(ts, t)
		at[t.Agent] = ts
	}

	return at, nil
}
