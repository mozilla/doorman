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
	os.Setenv("POLICIES", "sample.yaml")
	r, err := setupRouter()
	require.Nil(t, err)
	assert.Equal(t, 7, len(r.Routes()))
	assert.Equal(t, 3, len(r.RouterGroup.Handlers))

	os.Setenv("POLICIES", " \tsample.yaml")
	_, err = setupRouter()
	require.Nil(t, err)

	os.Unsetenv("POLICIES")
	_, err = setupRouter()
	require.NotNil(t, err)
	assert.Equal(t, "empty file \"policies.yaml\"", err.Error())
}

func TestSetupRouterBadPolicy(t *testing.T) {
	os.Setenv("POLICIES", "/tmp/unknown.yaml")
	defer os.Unsetenv("POLICIES")
	_, err := setupRouter()
	assert.NotNil(t, err)
}
