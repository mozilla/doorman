package api

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mozilla/doorman/doorman"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func performRequest(r http.Handler, method, path string, body io.Reader) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://sample.yaml")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func testJSONResponse(t *testing.T, url string, response interface{}) *httptest.ResponseRecorder {
	r := gin.New()
	SetupRoutes(r, doorman.NewDefaultLadon())
	w := performRequest(r, "GET", url, nil)

	assert.Equal(t, http.StatusOK, w.Code)
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
	// HTTP 404 if not found in current dir
	r := gin.New()
	SetupRoutes(r, doorman.NewDefaultLadon())
	w := performRequest(r, "GET", "/__version__", nil)
	assert.Equal(t, w.Code, http.StatusNotFound)

	// Copy to ./api/
	data, _ := ioutil.ReadFile("../version.json")
	ioutil.WriteFile("version.json", data, 0644)
	defer os.Remove("version.json")

	type Response struct {
		Commit string
	}
	var response Response
	testJSONResponse(t, "/__version__", &response)
	assert.Equal(t, response.Commit, "stub")
}

func TestOpenAPI(t *testing.T) {
	type Response struct {
		Openapi string
	}
	var response Response
	testJSONResponse(t, "/__api__", &response)

	assert.Equal(t, response.Openapi, "2.0.0")
}

func TestContribute(t *testing.T) {
	type Response struct {
		Name string
	}
	var response Response
	testJSONResponse(t, "/contribute.json", &response)

	assert.Equal(t, response.Name, "Doorman")
}
