package main

import (
	"net/http"
)

import (
	"github.com/gin-gonic/gin"
)

func lbHeartbeatHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"ok": true,
	})
}

func setupRoutes(r *gin.Engine) {
	r.GET("/__lbheartbeat__", lbHeartbeatHandler)
}

func main() {
	r := gin.Default()
	setupRoutes(r)
	r.Run() // listen and serve on 0.0.0.0:$PORT (:8080)
}
