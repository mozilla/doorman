package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var errorNumber = map[int]int{
	http.StatusOK:           0,
	http.StatusUnauthorized: 104,
	http.StatusForbidden:    121,
	http.StatusBadRequest:   109,
}

// MozLogger is a Gin middleware to log request summary following Mozilla Log format.
func MozLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Execute view.
		c.Next()

		// Stop timer
		end := time.Now()
		latency := end.Sub(start)

		log.WithFields(
			RequestLogFields(c.Request, c.Writer.Status(), latency),
		).Info("request.summary")
	}
}

// RequestLogFields returns the log fields from the request attributes.
func RequestLogFields(r *http.Request, statusCode int, latency time.Duration) log.Fields {
	path := r.URL.Path
	raw := r.URL.RawQuery
	if raw != "" {
		path = path + "?" + raw
	}
	// Error number.
	errno, defined := errorNumber[statusCode]
	if !defined {
		errno = 999
	}

	// See https://github.com/mozilla-services/go-mozlogrus/issues/5
	return log.Fields{
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
	}
}
