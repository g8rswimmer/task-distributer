package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/rs/xid"
)

func skillCount(db *sql.DB, skills []string) (int, error) {
	s := make([]string, len(skills))
	for idx, skill := range skills {
		s[idx] = fmt.Sprintf("'%s'", skill)
	}

	stmt := fmt.Sprintf(`SELECT COUNT(*) FROM SKILLS WHERE SKILL IN (%s)`, strings.Join(s, ","))
	fmt.Println(stmt)
	row := db.QueryRow(stmt)
	var count int
	err := row.Scan(&count)
	if err != nil {
		fmt.Println(err.Error())
		return -1, err
	}
	return count, nil
}
func priorityLevel(db *sql.DB, priority string) (int, error) {
	stmt := `SELECT PRIORITY_LEVEL FROM PRIORITIES WHERE PRIORITY = $1`
	row := db.QueryRow(stmt, priority)
	var level int
	err := row.Scan(&level)
	if err != nil {
		fmt.Println(err.Error())
		return -1, err
	}
	return level, nil
}

func matchingAgents(db *sql.DB, skills []string) ([]agent, error) {
	s := make([]string, len(skills))
	for idx, skill := range skills {
		s[idx] = fmt.Sprintf("'%s'", skill)
	}

	stmt := fmt.Sprintf(`SELECT AGENT FROM AGENTSKILLS WHERE SKILL IN (%s) GROUP BY AGENT HAVING COUNT(*) = %d`, strings.Join(s, ","), len(s))
	rows, err := db.Query(stmt)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, errors.New("no agents have the skills")
		}
		ids = append(ids, id)
	}

	fmt.Printf("%+v\n", ids)
	if len(ids) == 0 {
		return nil, errors.New("no agents have the skills")
	}

	return agentsdata(db, ids)

}

func agentsdata(db *sql.DB, ids []string) ([]agent, error) {
	i := make([]string, len(ids))
	for idx, id := range ids {
		i[idx] = fmt.Sprintf("'%s'", id)
	}

	stmt := fmt.Sprintf(`SELECT ID, FIRSTNAME, LASTNAME FROM AGENTS WHERE ID IN (%s)`, strings.Join(i, ","))
	rows, err := db.Query(stmt)
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

func createTask(db *sql.DB, ctp payload, agentID string) (task, error) {

	t := task{
		ID:        xid.New().String(),
		Name:      ctp.Name,
		Priorty:   ctp.Priorty,
		Skills:    ctp.Skills,
		Agent:     agentID,
		StartTime: time.Now(),
		Status:    "Assigned",
	}

	s := make([]string, len(t.Skills))
	for idx, skill := range t.Skills {
		s[idx] = fmt.Sprintf("'%s'", skill)
	}

	stmt := `
	INSERT INTO TASKS
	(ID, NAME, CREATEDATE, SKILLS, PRIORITY, STATUS, AGENT)
	VALUES
	('%s', '%s', now(), ARRAY[%s], '%s', '%s', '%s')
	`
	formattedStmt := fmt.Sprintf(stmt, t.ID, t.Name, strings.Join(s, ","), t.Priorty, t.Status, t.Agent)
	fmt.Println(formattedStmt)
	if _, err := db.Exec(formattedStmt); err != nil {
		return task{}, err
	}
	return t, nil
}

func agentTasks(db *sql.DB, agents []agent) (map[string][]task, error) {
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
	rows, err := db.Query(formattedStmt)
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
func recentAgent(db *sql.DB, aTasks map[string][]task, priorityLevel int) (string, error) {
	var ids []string
	for _, ts := range aTasks {
		for _, t := range ts {
			ids = append(ids, fmt.Sprintf("'%s'", t.Agent))
		}
	}

	stmt := `
	SELECT
	Agent
	FROM TASKS
	INNER JOIN PRIORITIES ON TASKS.priority = PRIORITIES.priority
	WHERE 
		Agent IN (%s)
	AND
		Status = 'Assigned'
	AND
		PRIORITIES.priority_level < %d
	ORDER BY Createdate DESC				 	
	`
	formattedStmt := fmt.Sprintf(stmt, strings.Join(ids, ","), priorityLevel)
	fmt.Println(formattedStmt)
	rows, err := db.Query(formattedStmt)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	defer rows.Close()
	agentID := ""
	for rows.Next() {
		if err := rows.Scan(&agentID); err != nil {
			return "", errors.New("unable to find an agent to assign the task")
		}
		break
	}

	return agentID, nil
}

func updateTaskStatus(db *sql.DB, id, status string) error {
	stmt := `
	UPDATE Tasks
	SET Status = '%s'
	WHERE 
		Id = '%s'
	`
	formattedStmt := fmt.Sprintf(stmt, status, id)
	fmt.Println(formattedStmt)
	_, err := db.Exec(formattedStmt)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil

}

func retrieveAgents(db *sql.DB) (map[string]agent, error) {
	stmt := `
	SELECT
	Id, FirstName, LastName
	FROM
	Agents
	`

	rows, err := db.Query(stmt)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer rows.Close()
	agentMap := map[string]agent{}

	for rows.Next() {
		var a agent
		if err := rows.Scan(&a.ID, &a.FirstName, &a.LastName); err != nil {
			return nil, errors.New("unable to retrieve agents")
		}
		agentMap[a.ID] = a
	}
	return agentMap, nil
}
func retrieveAgentTasks(db *sql.DB) ([]listAgentTasks, error) {
	agentMap, err := retrieveAgents(db)
	if err != nil {
		return nil, err
	}

	fmt.Println(agentMap)

	var ids []string
	agentTasks := map[string]listAgentTasks{}
	for id, a := range agentMap {
		ids = append(ids, fmt.Sprintf("'%s'", id))
		agentTasks[id] = listAgentTasks{
			agent: a,
		}
	}

	fmt.Println(agentTasks)

	stmt := `
	SELECT
	Id, Name, Agent, Priority, Skills, Createdate, Status, CompleteDate
	FROM Tasks
	WHERE 
		Agent IN (%s)
	AND
		Status = 'Assigned'	
	`

	formattedStmt := fmt.Sprintf(stmt, strings.Join(ids, ","))
	fmt.Println(formattedStmt)
	rows, err := db.Query(formattedStmt)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var t task
		var date pq.NullTime
		if err := rows.Scan(&t.ID, &t.Name, &t.Agent, &t.Priorty, pq.Array(&t.Skills), &t.StartTime, &t.Status, &date); err != nil {
			fmt.Println(err.Error())
			return nil, errors.New("unable to retrieve agent tasks")
		}
		if date.Valid {
			t.CompleteTime = date.Time
		}
		lat := agentTasks[t.Agent]
		lat.Tasks = append(lat.Tasks, t)
		agentTasks[t.Agent] = lat
		fmt.Println(t)
	}

	lats := make([]listAgentTasks, len(agentTasks))
	idx := 0
	for _, lat := range agentTasks {
		lats[idx] = lat
		idx++
	}
	return lats, nil
}
