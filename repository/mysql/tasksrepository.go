package mysql

import (
	"encoding/json"
	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	"goacs/models/tasks"
	"log"
	"time"
)

type TasksRepository struct {
	db *sqlx.DB
}

func NewTasksRepository(connection *sqlx.DB) TasksRepository {
	return TasksRepository{
		db: connection,
	}
}

func (t *TasksRepository) AddTask(task tasks.Task) {
	payload, _ := json.Marshal(task.Payload)
	if string(payload) == "null" {
		payload = []byte("{}")
	}

	dialect := goqu.Dialect("mysql")

	query, args, err := dialect.Insert("tasks").Prepared(true).
		Cols("for_name", "for_id", "event", "task", "not_before", "payload", "infinite", "created_at").
		Vals(goqu.Vals{
			task.ForName,
			task.ForID,
			task.Event,
			task.Task,
			task.NotBefore,
			payload,
			task.Infinite,
			task.CreatedAt,
		}).ToSQL()

	if err != nil {
		log.Println("ERROR", err.Error())
		log.Println("QUERY", query, args)
	}

	_, err = t.db.Exec(query, args...)

	if err != nil {
		log.Println("error AddTask ", err.Error())
	}
}

func (t *TasksRepository) UpdateTask(task tasks.Task) {
	payload, _ := json.Marshal(task.Payload)
	log.Println("PAYLOAD", payload)
	dialect := goqu.Dialect("mysql")

	query, args, _ := dialect.Update("tasks").Prepared(true).
		Set(goqu.Record{
			"for_name":   task.ForName,
			"for_id":     task.ForID,
			"event":      task.Event,
			"task":       task.Task,
			"not_before": task.NotBefore,
			"payload":    payload,
			"infinite":   task.Infinite,
		}).
		Where(goqu.Ex{
			"id": task.Id,
		}).
		ToSQL()

	_, err := t.db.Exec(query, args...)

	if err != nil {
		log.Println("UpdateTask Error", err.Error())
	}
}

func (t *TasksRepository) GetTask(id int64) tasks.Task {
	var globalTask tasks.Task
	err := t.db.Get(&globalTask, "SELECT * FROM tasks WHERE id=?", id)

	if err != nil {
		log.Println(err.Error())
	}

	return globalTask
}

func (t *TasksRepository) GetGlobalTasks() []tasks.Task {
	var globalTasks []tasks.Task
	err := t.db.Select(&globalTasks, "SELECT * FROM tasks WHERE for_name=? AND (done_at is null or infinite = true)", tasks.TASK_GLOBAL)

	if err != nil {
		log.Println(err.Error())
	}

	return globalTasks
}

func (t *TasksRepository) GetGlobalTask(id string) tasks.Task {
	var globalTask tasks.Task
	err := t.db.Get(&globalTask, "SELECT * FROM tasks WHERE for_name=? AND for_id=? AND (done_at is null or infinite = true)", tasks.TASK_GLOBAL, id)

	if err != nil {
		log.Println(err.Error())
	}

	return globalTask
}

func (t *TasksRepository) GetTasksForCPE(cpe_uuid string) []tasks.Task {
	var cpeTasks []tasks.Task
	err := t.db.Select(&cpeTasks, "SELECT * FROM tasks WHERE for_name=? AND for_id=? AND (done_at is null or infinite = true)", tasks.TASK_CPE, cpe_uuid)

	if err != nil {
		log.Println(err.Error())
	}

	return cpeTasks
}

func (t *TasksRepository) GetTasksForCPEWithoutDateCheck(cpe_uuid string) []tasks.Task {
	var cpeTasks []tasks.Task
	_ = t.db.Select(&cpeTasks, "SELECT * FROM tasks WHERE for_name=? AND for_id=? AND done_at is null", tasks.TASK_CPE, cpe_uuid)

	return cpeTasks
}

func (t *TasksRepository) GetAllTasksForCPE(cpe_uuid string) []tasks.Task {
	var cpeTasks []tasks.Task
	_ = t.db.Select(&cpeTasks, "SELECT * FROM tasks WHERE for_name=? AND for_id=?", tasks.TASK_CPE, cpe_uuid)

	return cpeTasks
}

func (t *TasksRepository) DoneTask(task_id int64) {
	_, _ = t.db.Exec("UPDATE tasks SET done_at = ? WHERE id = ?", time.Now(), task_id)

}
