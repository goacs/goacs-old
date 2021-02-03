package tasks

import (
	"goacs/acs/types"
	"gopkg.in/guregu/null.v4"
	"log"
	"time"
)

const (
	TASK_CPE    = "cpe"
	TASK_GLOBAL = "global"
)

const (
	GLOBAL_ID_NEW    = "new"
	GLOBAL_ID_INFORM = "inform"
)

const (
	RunScript      string = "RunScript"
	SendParameters        = "SendParameters"
	Reboot                = "Reboot"
	UploadFirmware        = "UploadFirmware"
)

type Task struct {
	Id              int64                        `json:"id" db:"id"`
	ForName         string                       `json:"for_name" db:"for_name"`
	ForID           string                       `json:"for_id" db:"for_id"`
	Event           string                       `json:"event" db:"event"`
	NotBefore       time.Time                    `json:"not_before" db:"not_before"`
	Task            string                       `json:"task" db:"task"`
	Script          string                       `json:"script" db:"script"`
	Infinite        bool                         `json:"infinite" db:"infinite"`
	CreatedAt       time.Time                    `json:"created_at" db:"created_at"`
	DoneAt          null.Time                    `json:"done_at" db:"done_at"`
	ParameterValues []types.ParameterValueStruct `json:"-"`
	ParameterInfo   []types.ParameterInfo        `json:"-"`
	NextLevel       bool                         `json:"-"`
}

func NewCPETask(cpe_uuid string) Task {
	return Task{
		ForName:   TASK_CPE,
		ForID:     cpe_uuid,
		NotBefore: time.Now(),
		CreatedAt: time.Now(),
	}
}

func NewGlobalTask(id string) Task {
	return Task{
		ForName:   TASK_GLOBAL,
		ForID:     id,
		Infinite:  true,
		NotBefore: time.Now(),
		CreatedAt: time.Now(),
	}
}

func FilterTasksByEvent(event string, tasksList []Task) []Task {
	var filteredTasks []Task
	for _, task := range tasksList {
		log.Println(task.Event, event)
		log.Println(task.Event == event)
		if task.Event == event {
			filteredTasks = append(filteredTasks, task)
		}
	}

	return filteredTasks
}
