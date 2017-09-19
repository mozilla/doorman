package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
	log "github.com/sirupsen/logrus"
	"go.mozilla.org/mozlogrus"
)


func MozLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Execute view.
		c.Next()

		// Stop timer
		end := time.Now()
		latency := end.Sub(start)

		// Request summary.
		r := c.Request
		path := r.URL.Path
		raw := r.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}
		// Error number.
		statusCode := c.Writer.Status()
		errno := 0
		if statusCode != http.StatusOK {
			errno = 999
		}

		// See https://github.com/mozilla-services/go-mozlogrus/issues/5
		log.WithFields(log.Fields{
			"remoteAddress":      r.RemoteAddr,
			"remoteAddressChain": [1]string{r.Header.Get("X-Forwarded-For")},
			"method":             r.Method,
			"agent":              r.Header.Get("User-Agent"),
			"code":               statusCode,
			"path":               path,
			"errno":              errno,
			"lang":               r.Header.Get("Accept-Language"),
			"t":                  latency / time.Millisecond,
			"uid":                nil,  // user id
			"rid":                nil,  // request id
			"service":            "",
			"context":            "",
		}).Info("request.summary")
	}
}

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

		body = yaml2json(body)

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
