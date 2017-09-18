package main

import (
	"io/ioutil"
	"net/http"
	"os"
)

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
)

func yaml2json(i interface{}) interface{} {
	// https://stackoverflow.com/a/40737676/141895
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = yaml2json(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = yaml2json(v)
		}
	}
	return i
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

func openApiHandler(c *gin.Context) {
	yamlFile, err := ioutil.ReadFile("openapi.yaml")
	if err != nil {
		c.AbortWithError(500, err)
		return
	}

	var body interface{}
	if err := yaml.Unmarshal(yamlFile, &body); err != nil {
		c.AbortWithError(500, err)
		return
	}

	body = yaml2json(body)

	c.JSON(http.StatusOK, body)
}

func setupRoutes(r *gin.Engine) {
	r.GET("/__lbheartbeat__", lbHeartbeatHandler)
	r.GET("/__heartbeat__", heartbeatHandler)
	r.GET("/__version__", versionHandler)
	r.GET("/__api__", openApiHandler)
}

func main() {
	r := gin.Default()
	setupRoutes(r)
	r.Run() // listen and serve on 0.0.0.0:$PORT (:8080)
}
