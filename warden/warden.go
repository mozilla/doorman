package warden

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Token struct {
	Subject  string `form:"subject" json:"subject" binding:"required"`
	Action   string `form:"action" json:"action" binding:"required"`
	Resource string `form:"resource" json:"resource" binding:"required"`
}

// Warden views to query the policies.
func SetupRoutes(r *gin.Engine) {

	// XXX: require Auth (currently hard-coded BasicAuth)
	authorized := r.Group("", gin.BasicAuth(gin.Accounts{
		"foo": "bar",
	}))

	authorized.POST("/allowed", allowedHandler)
}

func allowedHandler(c *gin.Context) {
	if c.Request.ContentLength == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Missing body",
		})
		return
	}

	var token Token
	if err := c.BindJSON(&token); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"allowed": false,
	})
}
