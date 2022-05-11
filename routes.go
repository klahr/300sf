package main

import (
	"git.sr.ht/~klahr/300sf/controllers"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	r.GET("/v1/user", controllers.User)
}
