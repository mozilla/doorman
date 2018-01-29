package handlers

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mozilla/doorman/config"
)

func TestMain(m *testing.M) {
	config.AddLoader(&config.FileLoader{})

	//Set Gin to Test Mode
	gin.SetMode(gin.TestMode)
	// Run the other tests
	os.Exit(m.Run())
}
