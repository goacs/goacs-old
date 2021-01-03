package controllers

import (
	"github.com/gin-gonic/gin"
	"goacs/http/request"
	"goacs/http/response"
	"goacs/models/tasks"
	"goacs/repository"
	"goacs/repository/mysql"
)

type AddGlobalTaskRequest struct {
	Event  string `json:"event" validate:"required"`
	Task   string `json:"task" validate:"required"`
	Script string `json:"script"`
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
	task.Script = globalTaskRequst.Script
	taskrepository.AddTask(task)
	response.ResponseData(ctx, "")
}
