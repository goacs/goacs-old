package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"goacs/models/user"
	"goacs/repository"
	"goacs/repository/mysql"
	"log"
)

type UserCreateRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func UserCreate(ctx *gin.Context) {
	var request UserCreateRequest
	err := json.NewDecoder(ctx.Request.Body).Decode(&request)
	if err != nil {
		log.Println("Error in req")
	}

	userModel := user.User{
		Username: request.Username,
		Password: user.EncryptPassword(request.Password),
		Email:    request.Email,
	}

	userRepository := mysql.NewUserRepository(repository.GetConnection())
	userInstance, err := userRepository.CreateUser(&userModel)
	log.Print(userModel, userInstance)
	json.NewEncoder(ctx.Writer).Encode(userInstance)
}
