package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.New()

	SetupRoutes(r)

	r.Run("localhost:8080")
}
