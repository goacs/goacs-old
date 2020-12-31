package http

import (
	"github.com/gin-gonic/gin"
	"goacs/http/controllers"
	"goacs/http/middleware/jwt"
	"goacs/lib"
)

func RegisterApiRoutes(gin *gin.Engine) {
	var env lib.Env
	gin.GET("/file/:filename", controllers.DownloadFile)
	apiGroup := gin.Group("/api")
	apiGroup.Use()
	apiGroup.POST("/auth/login", controllers.Login)

	apiGroup.Use(jwt.JWTAuthMiddleware(env.Get("JWT_SECRET", "")))
	{
		apiGroup.GET("/dashboard", controllers.GetDashboardData)
		apiGroup.POST("/user/create", controllers.UserCreate)

		apiGroup.GET("/config", controllers.GetConfig)
		apiGroup.POST("/config", controllers.SaveConfig)

		apiGroup.GET("/device", controllers.GetDevicesList)
		apiGroup.GET("/device/:uuid", controllers.GetDevice)
		apiGroup.DELETE("/device/:uuid", controllers.DeleteDevice)
		apiGroup.GET("/device/:uuid/kick", controllers.Kick)
		apiGroup.GET("/device/:uuid/parameters", controllers.GetDeviceParameters)
		apiGroup.POST("/device/:uuid/parameters", controllers.CreateParameter)
		apiGroup.PUT("/device/:uuid/parameters", controllers.UpdateParameter)
		apiGroup.DELETE("/device/:uuid/parameters", controllers.DeleteParameter)
		apiGroup.POST("/device/:uuid/addobject", controllers.AddObject)
		apiGroup.POST("/device/:uuid/getparametervalues", controllers.GetParameterValues)
		apiGroup.GET("/device/:uuid/tasks", controllers.GetDeviceQueuedTasks)
		apiGroup.POST("/device/:uuid/tasks", controllers.AddTaskForCPE)
		apiGroup.GET("/device/:uuid/templates", controllers.GetDeviceTemplates)
		apiGroup.POST("/device/:uuid/templates", controllers.AssignTemplateToDevice)
		apiGroup.DELETE("/device/:uuid/templates/:template_id", controllers.UnassignTemplateFromDevice)

		apiGroup.POST("/template", controllers.CreateTemplate)
		apiGroup.GET("/template", controllers.GetTemplatesList)
		apiGroup.GET("/template/:templateid", controllers.GetTemplate)
		apiGroup.POST("/template/:templateid/parameters", controllers.StoreTemplateParameter)
		apiGroup.GET("/template/:templateid/parameters", controllers.GetTemplateParameters)
		apiGroup.POST("/template/:templateid/parameters/:parameter_uuid", controllers.UpdateTemplateParameter)
		apiGroup.DELETE("/template/:templateid/parameters/:parameter_uuid", controllers.DeleteTemplateParameter)

		apiGroup.GET("/tasks", controllers.GetGlobalTasks)

		apiGroup.GET("/faults/today", controllers.GetTodayFaults)

		apiGroup.GET("/file", controllers.ListFiles)
		apiGroup.POST("/file", controllers.UploadFile)

	}
}
