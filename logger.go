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
	setLevelFromEnv(logrus.StandardLogger())

	summaryLog = logrus.Logger{
		Out:       os.Stdout,
		Formatter: &mozlogrus.MozLogFormatter{LoggerName: "iam", Type: "request.summary"},
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.InfoLevel,
	}
}

func setLevelFromEnv(logger *logrus.Logger) {
	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "fatal":
		logger.SetLevel(logrus.FatalLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	default:
		if gin.Mode() == gin.ReleaseMode {
			logger.SetLevel(logrus.InfoLevel)
		} else {
			logger.SetLevel(logrus.DebugLevel)
		}
	}
}

// HTTPLoggerMiddleware will log HTTP requests.
func HTTPLoggerMiddleware() gin.HandlerFunc {
	// For release mode, we log requests in JSON with Moz format.
	if gin.Mode() != gin.ReleaseMode {
		// Default Gin debug log.
		return gin.Logger()
	}
	mozlogrus.EnableFormatter(&mozlogrus.MozLogFormatter{LoggerName: "iam", Type: "app.log"})
	return MozLogger()
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
