package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/rs/xid"
)

// payload from the create task HTTP request
type payload struct {
	Name    string   `json:"name"`
	Skills  []string `json:"skills"`
	Priorty string   `json:"priority"`
}

func createPayload(body io.ReadCloser) (*payload, error) {
	decoder := json.NewDecoder(body)
	defer body.Close()
	var p payload
	err := decoder.Decode(&p)
	if err != nil {
		return nil, err
	}

	return &p, nil
}
func (p *payload) requiredFields() error {
	if p.Name == "" {
		return errors.New("name field must be present")
	}
	if p.Skills == nil {
		return errors.New("skills field must be present")
	}
	if p.Priorty == "" {
		return errors.New("priority field must be present")
	}
	return nil
}

func (p *payload) validateSkills(db *sql.DB) error {
	available, err := skillCount(destributerDb, p.Skills)
	if err != nil {
		return errors.New("unable to retrieve available skills")
	}
	if available != len(p.Skills) {
		return errors.New("task skills are not supported")
	}
	return nil
}

func (p *payload) validatePriority(db *sql.DB) error {
	level, err := priorityLevel(db, p.Priorty)
	if err != nil || level == -1 {
		return fmt.Errorf("task priority is not supported %s", p.Priorty)
	}
	return nil
}

func (p *payload) createTask(db *sql.DB, agentID string) (task, error) {

	t := task{
		ID:        xid.New().String(),
		Name:      p.Name,
		Priorty:   p.Priorty,
		Skills:    p.Skills,
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
	if _, err := db.Exec(formattedStmt); err != nil {
		return task{}, err
	}
	return t, nil
}

// task that is distributed to an agent
type task struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Skills        []string `json:"skills"`
	Priorty       string   `json:"priority"`
	priorityLevel int
	Status        string    `json:"status"`
	StartTime     time.Time `json:"start_time"`
	CompleteTime  time.Time `json:"complete_time,omitempty"`
	Agent         string    `json:"assigned_agent"`
	db            *sql.DB
}

func (t *task) assignTask(p payload) error {
	skilledAgents, err := matchingAgents(t.db, p.Skills)
	if err != nil {
		return err
	}
	agents := agents{
		db: t.db,
	}
	ats, err := agents.tasks(skilledAgents)
	if err != nil {
		return err
	}
	level, err := priorityLevel(t.db, p.Priorty)
	if err != nil {
		return err
	}
	if level < 0 {
		return errors.New("unable to find an agent to assign the task")
	}

	for _, skilledAgent := range skilledAgents {
		if tsks, has := ats[skilledAgent.ID]; has {
			for _, tsk := range tsks {
				if tsk.priorityLevel >= level {
					delete(ats, skilledAgent.ID)
				}
			}
		} else {
			return t.insert(p, skilledAgent.ID)
		}
	}

	if len(ats) == 0 {
		return errors.New("unable to find an agent to assign the task")
	}
	id, err := recentAgent(t.db, ats, level)
	if err != nil {
		return err
	}

	if id == "" {
		return errors.New("unable to find an agent to assign the task")
	}
	return t.insert(p, id)
}
func (t *task) insert(ctp payload, agentID string) error {

	t.ID = xid.New().String()
	t.Name = ctp.Name
	t.Priorty = ctp.Priorty
	t.Skills = ctp.Skills
	t.Agent = agentID
	t.StartTime = time.Now()
	t.Status = "Assigned"

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
	if _, err := t.db.Exec(formattedStmt); err != nil {
		return err
	}
	return nil
}

func (t *task) retrieve(id string) error {
	stmt := `
	SELECT
	Id, Agent, Priority, Skills, Createdate, Status, CompleteDate
	FROM Tasks
	WHERE 
		Id = '%s'
	`
	formattedStmt := fmt.Sprintf(stmt, id)
	fmt.Println(formattedStmt)
	rows, err := t.db.Query(formattedStmt)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	defer rows.Close()
	var tsk task
	for rows.Next() {
		var date pq.NullTime
		if err := rows.Scan(&tsk.ID, &tsk.Agent, &tsk.Priorty, pq.Array(&tsk.Skills), &tsk.StartTime, &tsk.Status, &date); err != nil {
			fmt.Println(err.Error())
			return fmt.Errorf("unable to find task %s", id)
		}
		if date.Valid {
			tsk.CompleteTime = date.Time
		}
		break
	}

	if tsk.ID == "" {
		return fmt.Errorf("unable to find task %s", id)
	}

	t.ID = tsk.ID
	t.Agent = tsk.Agent
	t.Priorty = tsk.Priorty
	t.StartTime = tsk.StartTime
	t.Status = tsk.Status
	t.Skills = tsk.Skills
	t.CompleteTime = tsk.CompleteTime

	return nil
}
