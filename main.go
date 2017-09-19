package main

import (
	"io/ioutil"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.mozilla.org/mozlogrus"
	"gopkg.in/yaml.v2"
)


func YAMLAsJSONHandler(filename string) gin.HandlerFunc {
	return func(c *gin.Context) {
		yamlFile, err := ioutil.ReadFile(filename)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}

		var body interface{}
		if err := yaml.Unmarshal(yamlFile, &body); err != nil {
			c.AbortWithError(500, err)
			return
		}

		body = Yaml2JSON(body)

		c.JSON(http.StatusOK, body)
	}
}

func lbHeartbeatHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"ok": true,
	})
}

func heartbeatHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

func versionHandler(c *gin.Context) {
	versionFile := os.Getenv("VERSION_FILE")
	if versionFile == "" {
		versionFile = "version.json"
	}
	c.File(versionFile)
}

func SetupRouter() *gin.Engine {
	r := gin.New()
	// Crash free (turns errors into 5XX).
	r.Use(gin.Recovery())

	// Setup logging.
	if gin.Mode() == gin.ReleaseMode {
		// See https://github.com/mozilla-services/go-mozlogrus/issues/2#issuecomment-330495098
		r.Use(MozLogger())
		mozlogrus.Enable("iam")
	} else {
		r.Use(gin.Logger())
	}

	// Anonymous utilities views.
	r.GET("/__lbheartbeat__", lbHeartbeatHandler)
	r.GET("/__heartbeat__", heartbeatHandler)
	r.GET("/__version__", versionHandler)
	r.GET("/__api__", YAMLAsJSONHandler("openapi.yaml"))
	r.GET("/contribute.json", YAMLAsJSONHandler("contribute.yaml"))

	return r
}

func main() {
	r := SetupRouter()
	r.Run() // listen and serve on 0.0.0.0:$PORT (:8080)
}
