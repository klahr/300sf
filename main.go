package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.New()
	r.Run("localhost:8080")
}
