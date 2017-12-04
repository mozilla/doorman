package main

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

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
		assert.Equal(t, test.level, levelFromEnv())
	}
}

func TestSources(t *testing.T) {
	os.Setenv("POLICIES", " \tsample.yaml")
	defer os.Unsetenv("POLICIES")
	assert.Equal(t, []string{"sample.yaml"}, sources())
}
