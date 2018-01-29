package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mozilla/doorman/config"
	"github.com/mozilla/doorman/doorman"
)

func reloadHandler(sources []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Load files (from folders, files, Github, etc.)
		configs, err := config.Load(sources)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}

		// Load into Doorman.
		d := c.MustGet(DoormanContextKey).(doorman.Doorman)

		if err := d.LoadPolicies(configs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "",
		})
	}
}
