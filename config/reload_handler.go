package config

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mozilla/doorman/doorman"
)

// SetupRoutes adds the reload config view.
func SetupRoutes(r *gin.Engine, sources []string) {
	r.POST("/__reload__", reloadHandler(sources))
}

func reloadHandler(sources []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Load files (from folders, files, Github, etc.)
		configs, err := Load(sources)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}

		// Load into Doorman.
		doorman := c.MustGet(doorman.DoormanContextKey).(doorman.Doorman)

		if err := doorman.LoadPolicies(configs); err != nil {
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
