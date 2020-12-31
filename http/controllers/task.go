package controllers

import (
	"github.com/gin-gonic/gin"
	"goacs/http/request"
	"goacs/http/response"
	"goacs/models/tasks"
	"goacs/repository"
	"goacs/repository/mysql"
)

type AddTaskForCPERequest struct {
	Event  string `json:"event" validate:"required"`
	Task   string `json:"task" validate:"required"`
	Script string `json:"script"`
}

func GetGlobalTasks(ctx *gin.Context) {
	taskrepository := mysql.NewTasksRepository(repository.GetConnection())
	response.ResponseData(ctx, taskrepository.GetGlobalTasks())
}

func AddTaskForCPE(ctx *gin.Context) {
	var addTaskRequest AddTaskForCPERequest
	_ = ctx.BindJSON(&addTaskRequest)

	validator := request.NewApiValidator(ctx, addTaskRequest)
	verr := validator.Validate()

	if verr != nil {
		response.ResponseValidationErrors(ctx, validator)
		return
	}

	cperepository := mysql.NewCPERepository(repository.GetConnection())
	taskrepository := mysql.NewTasksRepository(repository.GetConnection())
	cpeModel, err := getCPEFromContext(ctx, cperepository)

	if err != nil {
		return
	}

	task := tasks.NewCPETask(cpeModel.UUID)
	task.Event = addTaskRequest.Event
	task.Task = addTaskRequest.Task
	task.Script = addTaskRequest.Script

	taskrepository.AddTask(task)
	response.ResponseData(ctx, "")
}
