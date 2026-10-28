package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NotImplemented(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, GeneralErrorResponse{
		Message: "not implemented",
		Code:    ErrorInvalidRequest,
	})
}

type SingleItemDataResponse struct {
	Data interface{} `json:"data"`
}

type ListDataResponse struct {
	Data []interface{} `json:"data"`
}

type GeneralSuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type GeneralErrorResponse struct {
	Message string `json:"message"`
	Code    int64  `json:"code"`
}

type ValidationErrorResponse struct {
	Message string `json:"message"`
	Errors  []struct {
		Field   string `json:"field"`
		Message string `json:"message"`
	}
}
