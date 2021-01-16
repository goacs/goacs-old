package controllers

import (
	"github.com/gin-gonic/gin"
	"goacs/acs/types"
	"goacs/http/request"
	"goacs/http/response"
	"goacs/models/templates"
	"goacs/repository"
	"goacs/repository/mysql"
	"strconv"
)

type TemplateListResponse struct {
	templates.Template
	ParameterCount int64 `json:"parameter_count"`
}

type TemplateStoreRequest struct {
	Name string `json:"name" validate:"required"`
}

type TemplateParameterStoreRequest struct {
	TemplateId int64      `json:"template_id" validate:"required"`
	Name       string     `json:"name" validate:"required"`
	Value      string     `json:"value"`
	Flag       types.Flag `json:"flag" validate:"required"`
}

type TemplateParameterUpdateRequest struct {
	TemplateId    int64      `json:"template_id" validate:"required"`
	ParameterUUID string     `json:"parameter_uuid" validate:"required"`
	Name          string     `json:"name" validate:"required"`
	Value         string     `json:"value"`
	Flag          types.Flag `json:"flag" validate:"required"`
}

type TemplateParameterDeleteRequest struct {
	TemplateId    int64  `json:"template_id" validate:"required"`
	ParameterUUID string `json:"parameter_uuid" validate:"required"`
}

func CreateTemplate(ctx *gin.Context) {
	var templateStoreRequest TemplateStoreRequest
	_ = ctx.ShouldBindJSON(&templateStoreRequest)

	validator := request.NewApiValidator(ctx, templateStoreRequest)

	err := validator.Validate()

	if err != nil {
		response.ResponseValidationErrors(ctx, validator)
		return
	}

	templatesrepository := mysql.NewTemplateRepository(repository.GetConnection())
	templatesrepository.CreateTemplate(&templates.Template{
		Name: templateStoreRequest.Name,
	})

	response.ResponseData(ctx, "")
}

func GetTemplatesList(ctx *gin.Context) {
	paginatorRequest := repository.PaginatorRequestFromContext(ctx)
	templatesrepository := mysql.NewTemplateRepository(repository.GetConnection())
	templatesList, total := templatesrepository.List(paginatorRequest)
	templatesList = templatesrepository.HydrateTemplatesParameters(templatesList)

	var templateResponse []TemplateListResponse

	for _, template := range templatesList {
		count := int64(len(template.Parameters))
		template.Parameters = []templates.TemplateParameter{}
		templateResponse = append(templateResponse, TemplateListResponse{
			Template:       template,
			ParameterCount: count,
		})
	}

	responseData := repository.NewPaginatorResponse(paginatorRequest, total, templateResponse)
	response.ResponsePaginatior(ctx, responseData)
}

func GetTemplate(ctx *gin.Context) {
	templateId, err := strconv.Atoi(ctx.Param("templateid"))

	if err != nil {
		response.ResponseError(ctx, 404, "", "")
		return
	}

	templatesrepository := mysql.NewTemplateRepository(repository.GetConnection())
	template, err := templatesrepository.Find(int64(templateId))

	if err != nil {
		response.ResponseError(ctx, 404, "", "")
		return
	}

	response.ResponseData(ctx, template)
}

func GetTemplateParameters(ctx *gin.Context) {
	templateId, err := strconv.Atoi(ctx.Param("templateid"))

	paginatorRequest := repository.PaginatorRequestFromContext(ctx)
	templatesrepository := mysql.NewTemplateRepository(repository.GetConnection())
	template, err := templatesrepository.Find(int64(templateId))

	if err == nil {
		parameters, total := templatesrepository.ListTemplateParameters(template, paginatorRequest)
		responseData := repository.NewPaginatorResponse(paginatorRequest, total, parameters)
		response.ResponsePaginatior(ctx, responseData)
		return
	}
}

func UpdateTemplateParameter(ctx *gin.Context) {
	var templatePURequest TemplateParameterUpdateRequest
	parameterId := ctx.Param("parameter_uuid")
	templateId, _ := strconv.Atoi(ctx.Param("templateid"))

	templatePURequest.ParameterUUID = parameterId
	templatePURequest.TemplateId = int64(templateId)
	_ = ctx.ShouldBindJSON(&templatePURequest)

	validator := request.NewApiValidator(ctx, templatePURequest)

	err := validator.Validate()

	if err != nil {
		response.ResponseValidationErrors(ctx, validator)
		return
	}

	templatesRepository := mysql.NewTemplateRepository(repository.GetConnection())
	err = templatesRepository.UpdateParameter(templatePURequest.ParameterUUID,
		types.ParameterValueStruct{
			Name: templatePURequest.Name,
			ValueStruct: types.ValueStruct{
				Value: templatePURequest.Value,
				Type:  "",
			},
			Flag: templatePURequest.Flag,
		},
	)

	if err != nil {
		response.Response500(ctx, "Error", err)
		return
	}

	response.ResponseData(ctx, "")

}

func StoreTemplateParameter(ctx *gin.Context) {
	var templatePSRequest TemplateParameterStoreRequest
	templateId, _ := strconv.Atoi(ctx.Param("templateid"))

	templatePSRequest.TemplateId = int64(templateId)
	_ = ctx.ShouldBindJSON(&templatePSRequest)

	validator := request.NewApiValidator(ctx, templatePSRequest)

	err := validator.Validate()

	if err != nil {
		response.ResponseValidationErrors(ctx, validator)
		return
	}

	templatesRepository := mysql.NewTemplateRepository(repository.GetConnection())
	err = templatesRepository.CreateParameter(templatePSRequest.TemplateId,
		types.ParameterValueStruct{
			Name: templatePSRequest.Name,
			ValueStruct: types.ValueStruct{
				Value: templatePSRequest.Value,
				Type:  "",
			},
			Flag: templatePSRequest.Flag,
		},
	)

	if err != nil {
		response.Response500(ctx, "Error", err.Error())
		return
	}

	response.ResponseData(ctx, "")
}

func DeleteTemplateParameter(ctx *gin.Context) {
	var templatePDRequest TemplateParameterDeleteRequest
	parameterId := ctx.Param("parameter_uuid")
	templateId, _ := strconv.Atoi(ctx.Param("templateid"))

	templatePDRequest.TemplateId = int64(templateId)
	templatePDRequest.ParameterUUID = parameterId
	_ = ctx.ShouldBindJSON(&templatePDRequest)
	validator := request.NewApiValidator(ctx, templatePDRequest)

	err := validator.Validate()

	if err != nil {
		response.ResponseValidationErrors(ctx, validator)
		return
	}

	templatesRepository := mysql.NewTemplateRepository(repository.GetConnection())

	err = templatesRepository.DeleteParameter(templatePDRequest.ParameterUUID, templatePDRequest.TemplateId)

	if err != nil {
		response.Response500(ctx, "Error", err.Error())
		return
	}

	response.ResponseData(ctx, "")

}
