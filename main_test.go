package main

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
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
	_, err := setupRouter()
	require.NotNil(t, err)
	assert.Equal(t, "empty file \"policies.yaml\"", err.Error())

	config.Sources = []string{"sample.yaml"}
	defer func() {
		config.Sources = []string{DefaultPoliciesFilename}
	}()

	r, err := setupRouter()
	require.Nil(t, err)
	assert.Equal(t, 7, len(r.Routes()))
	assert.Equal(t, 3, len(r.RouterGroup.Handlers))
}
