package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mozilla.org/mozlogrus"
)

var summaryLog logrus.Logger

var errorNumber = map[int]int{
	http.StatusOK:           0,
	http.StatusUnauthorized: 104,
	http.StatusForbidden:    121,
	http.StatusBadRequest:   109,
}

func init() {
	summaryLog = logrus.Logger{
		Out:       os.Stdout,
		Formatter: &mozlogrus.MozLogFormatter{LoggerName: "doorman", Type: "request.summary"},
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.InfoLevel,
	}
}

func setupLogging() {
	logrus.StandardLogger().SetLevel(settings.LogLevel)
	if gin.Mode() == gin.ReleaseMode {
		mozlogrus.EnableFormatter(&mozlogrus.MozLogFormatter{LoggerName: "doorman", Type: "app.log"})
	}
}

// HTTPLoggerMiddleware will log HTTP requests.
func HTTPLoggerMiddleware() gin.HandlerFunc {
	// For release mode, we log requests in JSON with Moz format.
	if gin.Mode() != gin.ReleaseMode {
		// Default Gin debug log.
		return gin.Logger()
	}
	return RequestSummaryLogger()
}

// RequestSummaryLogger is a Gin middleware to log request summary following Mozilla Log format.
func RequestSummaryLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Execute view.
		c.Next()

		// Stop timer
		end := time.Now()
		latency := end.Sub(start)

		summaryLog.WithFields(
			RequestLogFields(c.Request, c.Writer.Status(), latency),
		).Info("")
	}
}

// RequestLogFields returns the log fields from the request attributes.
func RequestLogFields(r *http.Request, statusCode int, latency time.Duration) logrus.Fields {
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
	return logrus.Fields{
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
