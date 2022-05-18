package main

import (
	"git.sr.ht/~klahr/300sf/controllers"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// Users
	r.POST("/v1/users", controllers.User)
}
