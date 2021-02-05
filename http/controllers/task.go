package controllers

import (
	"github.com/gin-gonic/gin"
	"goacs/http/request"
	"goacs/http/response"
	"goacs/models/tasks"
	"goacs/repository"
	"goacs/repository/mysql"
	"strconv"
)

type AddGlobalTaskRequest struct {
	Event   string `json:"event" validate:"required"`
	Task    string `json:"task" validate:"required"`
	Payload string `json:"payload"`
}

type UpdateGlobalTaskRequest struct {
	TaskId int64 `validate:"required"`
	AddGlobalTaskRequest
}

func GetGlobalTasks(ctx *gin.Context) {
	taskrepository := mysql.NewTasksRepository(repository.GetConnection())
	response.ResponseData(ctx, taskrepository.GetGlobalTasks())
}

func AddGlobalTask(ctx *gin.Context) {
	var globalTaskRequst AddGlobalTaskRequest
	_ = ctx.BindJSON(&globalTaskRequst)

	validator := request.NewApiValidator(ctx, globalTaskRequst)
	verr := validator.Validate()

	if verr != nil {
		response.ResponseValidationErrors(ctx, validator)
		return
	}

	taskrepository := mysql.NewTasksRepository(repository.GetConnection())

	task := tasks.NewGlobalTask(tasks.GLOBAL_ID_NEW)
	task.Event = globalTaskRequst.Event
	task.Task = globalTaskRequst.Task
	//task.Payload = globalTaskRequst.Payload
	taskrepository.AddTask(task)
	response.ResponseData(ctx, "")
}

func UpdateGlobalTask(ctx *gin.Context) {
	var globalTaskRequst UpdateGlobalTaskRequest
	_ = ctx.BindJSON(&globalTaskRequst)

	globalTaskRequst.TaskId, _ = strconv.ParseInt(ctx.Param("taskid"), 10, 64)

	validator := request.NewApiValidator(ctx, globalTaskRequst)
	verr := validator.Validate()

	if verr != nil {
		response.ResponseValidationErrors(ctx, validator)
		return
	}

	taskrepository := mysql.NewTasksRepository(repository.GetConnection())
	task := taskrepository.GetTask(globalTaskRequst.TaskId)

	task.Event = globalTaskRequst.Event
	task.Task = globalTaskRequst.Task
	//task.Payload = globalTaskRequst.Payload

	taskrepository.UpdateTask(task)
	response.ResponseData(ctx, "")
}
