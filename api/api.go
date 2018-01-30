package api

import (
	"github.com/gin-gonic/gin"
	"github.com/mozilla/doorman/doorman"
)

// SetupRoutes adds HTTP endpoints to the gin.Engine.
func SetupRoutes(r *gin.Engine, d doorman.Doorman) {
	r.Use(ContextMiddleware(d))

	a := r.Group("")
	a.Use(AuthnMiddleware(d))
	a.POST("/allowed", allowedHandler)

	sources := d.ConfigSources()
	r.POST("/__reload__", reloadHandler(sources))

	r.GET("/__lbheartbeat__", lbHeartbeatHandler)
	r.GET("/__heartbeat__", heartbeatHandler)
	r.GET("/__version__", versionHandler)
	r.GET("/__api__", YAMLAsJSONHandler("api/openapi.yaml"))
	r.GET("/contribute.json", YAMLAsJSONHandler("api/contribute.yaml"))
}
