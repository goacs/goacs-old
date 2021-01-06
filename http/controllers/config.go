package controllers

import (
	"github.com/gin-gonic/gin"
	"goacs/http/request"
	"goacs/http/response"
	"goacs/repository"
	"goacs/repository/mysql"
	"log"
)

type SaveConfigRequest struct {
	Config map[string]string `json:"config" validation:"required"`
}

func SaveConfig(ctx *gin.Context) {
	var saveConfigRequest SaveConfigRequest
	_ = ctx.BindJSON(&saveConfigRequest)
	validator := request.NewApiValidator(ctx, saveConfigRequest)
	verr := validator.Validate()

	if verr != nil {
		response.ResponseValidationErrors(ctx, validator)
		return
	}

	configRepository := mysql.NewConfigRepository(repository.GetConnection())

	for key, value := range saveConfigRequest.Config {
		configRepository.SetValue(key, value)
	}

	response.ResponseData(ctx, "")
}

func GetConfig(ctx *gin.Context) {
	configRepository := mysql.NewConfigRepository(repository.GetConnection())
	values := configRepository.GetValues()
	log.Println(values)
	response.ResponseData(ctx, values)
}
