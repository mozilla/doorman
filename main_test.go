package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	//Set Gin to Test Mode
	gin.SetMode(gin.TestMode)
	// Run the other tests
	os.Exit(m.Run())
}

func TestSetupRouter(t *testing.T) {
	r := setupRouter()
	assert.Equal(t, 6, len(r.Routes()))
	assert.Equal(t, 3, len(r.RouterGroup.Handlers))

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
