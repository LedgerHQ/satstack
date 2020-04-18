package main

import (
	"ledger-sat-stack/httpd/controllers"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.GET("/ping", controllers.PingGet)
	r.Run()
}
