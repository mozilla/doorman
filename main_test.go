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
	os.Setenv("POLICIES_FILES", "sample.yaml")
	defer os.Unsetenv("POLICIES_FILES")
	r, err := setupRouter()
	require.Nil(t, err)
	assert.Equal(t, 6, len(r.Routes()))
	assert.Equal(t, 3, len(r.RouterGroup.Handlers))

	os.Setenv("POLICIES_FILES", " \tsample.yaml")
	defer os.Unsetenv("POLICIES_FILES")
	_, err = setupRouter()
	require.Nil(t, err)
}

func TestSetupRouterBadPolicy(t *testing.T) {
	os.Setenv("POLICIES_FILES", "/tmp/unknown.yaml")
	defer os.Unsetenv("POLICIES_FILES")
	_, err := setupRouter()
	assert.NotNil(t, err)
}
