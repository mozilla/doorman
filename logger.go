package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
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
			"uid":                nil, // user id
			"rid":                nil, // request id
			"service":            "",
			"context":            "",
		}).Info("request.summary")
	}
}
