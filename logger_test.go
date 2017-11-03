package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"

	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestLoggerMiddleware(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("GET", "/get", nil)
	handler := MozLogger()

	var buf bytes.Buffer
	summaryLog.Out = &buf

	handler(c)

	summaryLog.Out = os.Stdout

	assert.Contains(t, buf.String(), "\"errno\":0")
}

func TestRequestLogFields(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	fields := RequestLogFields(r, 200, time.Duration(100))
	assert.Equal(t, 200, fields["code"])

	// Errno
	fields = RequestLogFields(r, 500, time.Duration(100))
	assert.Equal(t, 999, fields["errno"])
	fields = RequestLogFields(r, 400, time.Duration(100))
	assert.Equal(t, 109, fields["errno"])
	fields = RequestLogFields(r, 401, time.Duration(100))
	assert.Equal(t, 104, fields["errno"])
	fields = RequestLogFields(r, 403, time.Duration(100))
	assert.Equal(t, 121, fields["errno"])

	r, _ = http.NewRequest("POST", "/diff?w=1", nil)
	fields = RequestLogFields(r, 200, time.Duration(100))
	assert.Equal(t, "/diff?w=1", fields["path"])
}

func TestEnvLogLevel(t *testing.T) {
	logger := logrus.StandardLogger()
	var cases = []struct {
		mode  string
		env   string
		level logrus.Level
	}{
		{gin.DebugMode, "fatal", logrus.FatalLevel},
		{gin.DebugMode, "error", logrus.ErrorLevel},
		{gin.DebugMode, "warn", logrus.WarnLevel},
		{gin.DebugMode, "debug", logrus.DebugLevel},
		{gin.DebugMode, "info", logrus.DebugLevel},
		{gin.ReleaseMode, "info", logrus.InfoLevel},
	}
	defer gin.SetMode(gin.TestMode)
	defer os.Unsetenv("LOG_LEVEL")
	for _, test := range cases {
		gin.SetMode(test.mode)
		os.Setenv("LOG_LEVEL", test.env)
		setLevelFromEnv(logger)
		assert.Equal(t, test.level, logger.Level)
	}
}

func TestSetupRouterRelease(t *testing.T) {
	// In release mode, we enable MozLogger middleware.
	gin.SetMode(gin.ReleaseMode)
	defer gin.SetMode(gin.TestMode)
	setupRouter()

	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	defer logrus.SetOutput(os.Stdout)

	logrus.Info("Haha")

	assert.Contains(t, buf.String(), "\"msg\":\"Haha\"")
}
