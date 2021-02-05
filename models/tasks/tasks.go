package tasks

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"goacs/acs/types"
	"gopkg.in/guregu/null.v4"
	"log"
	"regexp"
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

type TaskPayload map[string]interface{}

type Task struct {
	Id              int64                        `json:"id" db:"id"`
	ForName         string                       `json:"for_name" db:"for_name"`
	ForID           string                       `json:"for_id" db:"for_id"`
	Event           string                       `json:"event" db:"event"`
	NotBefore       time.Time                    `json:"not_before" db:"not_before"`
	Task            string                       `json:"task" db:"task"`
	Payload         TaskPayload                  `json:"payload" db:"payload"`
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

func (task *Task) AsScript(script string) {
	task.Task = RunScript
	task.Payload = TaskPayload{
		"script": script,
	}
}

func (task *Task) AsUploadFirmware(filename string, filetype string) {
	task.Task = UploadFirmware
	task.Payload = TaskPayload{
		"filename": filename,
		"filetype": filetype,
	}
}

func (i *TaskPayload) Value() (driver.Value, error) {
	return json.Marshal(i)
}

func (i *TaskPayload) Scan(src interface{}) (err error) {

	switch src.(type) {
	case []uint8:
		src := string(src.([]byte))
		re := regexp.MustCompile(`\r?\n`)
		src = re.ReplaceAllString(src, " ")
		err = json.Unmarshal([]byte(src), i)
	default:
		err = errors.New("Invalid payload")
	}

	return
}

//
//func (i *TaskPayload) UnmarshalJSON(b []byte) error {
//	err := json.Unmarshal(b, i)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}
//
//func (i *TaskPayload) MarshalJSON() ([]byte, error) {
//	jsonstring, err := json.Marshal(i)
//	if err != nil {
//		return nil, err
//	}
//	return jsonstring, nil
//}
