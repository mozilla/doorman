package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	//Set Gin to Test Mode
	gin.SetMode(gin.TestMode)
	// Run the other tests
	os.Exit(m.Run())
}

func TestSetupRouter(t *testing.T) {
	os.Setenv("POLICIES_FILE", "sample.yaml")
	defer os.Unsetenv("POLICIES_FILE")
	r, err := setupRouter()
	require.Nil(t, err)
	assert.Equal(t, 6, len(r.Routes()))
	assert.Equal(t, 3, len(r.RouterGroup.Handlers))
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

func TestSetupRouterBadPolicy(t *testing.T) {
	os.Setenv("POLICIES_FILE", "/tmp/unknown.yaml")
	defer os.Unsetenv("POLICIES_FILE")
	_, err := setupRouter()
	assert.NotNil(t, err)
}

func TestEnvLogLevel(t *testing.T) {
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
		setLogLevel()
		assert.Equal(t, test.level, logrus.StandardLogger().Level)
	}
}
