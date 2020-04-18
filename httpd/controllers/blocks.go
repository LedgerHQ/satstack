package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func PingGet(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
