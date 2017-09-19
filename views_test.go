package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func performRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func testJSONResponse(t *testing.T, url string, response interface{}) *httptest.ResponseRecorder {
	r := SetupRouter()
	w := performRequest(r, "GET", url)

	assert.Equal(t, w.Code, http.StatusOK)
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.Nil(t, err)

	return w
}

func TestLBHeartbeat(t *testing.T) {
	type Response struct {
		Ok bool
	}
	var response Response
	testJSONResponse(t, "/__lbheartbeat__", &response)

	assert.True(t, response.Ok)
}

func TestHeartbeat(t *testing.T) {
	type Response struct {
	}
	var response Response
	testJSONResponse(t, "/__heartbeat__", &response)
}

func TestVersion(t *testing.T) {
	type Response struct {
		Commit string
	}
	var response Response
	testJSONResponse(t, "/__version__", &response)

	assert.Equal(t, response.Commit, "stub")
}

func TestVersionMissing(t *testing.T) {
	os.Setenv("VERSION_FILE", "/tmp/missing.json")

	r := SetupRouter()
	w := performRequest(r, "GET", "/__version__")

	assert.Equal(t, w.Code, http.StatusNotFound)
}

func TestOpenAPI(t *testing.T) {
	type Response struct {
		Swagger string
	}
	var response Response
	testJSONResponse(t, "/__api__", &response)

	assert.Equal(t, response.Swagger, "2.0")
}

func TestContribute(t *testing.T) {
	type Response struct {
		Name string
	}
	var response Response
	testJSONResponse(t, "/contribute.json", &response)

	assert.Equal(t, response.Name, "IAM")
}
