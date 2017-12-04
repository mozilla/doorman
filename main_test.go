package main

import (
	"io/ioutil"
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
	// Empty file.
	_, err := setupRouter()
	require.NotNil(t, err)
	assert.Equal(t, "empty file \"policies.yaml\"", err.Error())

	// Bad definition (unknown condition type).
	tmpfile, _ := ioutil.TempFile("", "")
	tmpfile.Write([]byte(`
service: a
policies:
  -
    id: "1"
    action: update
    conditions:
      owner:
        type: fantastic
`))
	settings.Sources = []string{tmpfile.Name()}
	_, err = setupRouter()
	require.NotNil(t, err)
	assert.Equal(t, "unknown condition type fantastic", err.Error())

	defer func() {
		os.Remove(tmpfile.Name()) // clean up
		settings.Sources = []string{DefaultPoliciesFilename}
	}()

	// Sample file.
	settings.Sources = []string{"sample.yaml"}
	r, err := setupRouter()
	require.Nil(t, err)
	assert.Equal(t, 7, len(r.Routes()))
	assert.Equal(t, 3, len(r.RouterGroup.Handlers))
}
