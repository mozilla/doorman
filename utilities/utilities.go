package utilities

import (
	"io/ioutil"
	"net/http"
	"path/filepath"
	"os"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
)

func SetupRoutes(r *gin.Engine) {
	// Anonymous utilities views.
	r.GET("/__lbheartbeat__", lbHeartbeatHandler)
	r.GET("/__heartbeat__", heartbeatHandler)
	r.GET("/__version__", versionHandler)
	r.GET("/__api__", YAMLAsJSONHandler("openapi.yaml"))
	r.GET("/contribute.json", YAMLAsJSONHandler("contribute.yaml"))
}

func Yaml2JSON(i interface{}) interface{} {
	// https://stackoverflow.com/a/40737676/141895
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = Yaml2JSON(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = Yaml2JSON(v)
		}
	}
	return i
}

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
		// Look in current working directory.
		here, _ := os.Getwd()
		versionFile = filepath.Join(here, "version.json")
	}
	c.File(versionFile)
}
