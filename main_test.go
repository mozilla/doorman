package main

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
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
}
