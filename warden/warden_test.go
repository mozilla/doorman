package warden

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ory/ladon"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Response struct {
	Allowed bool
}

type ErrorResponse struct {
	Message string
}

func TestMain(m *testing.M) {
	// Use sample file.
	os.Setenv("POLICIES_FILE", "../sample.yaml")
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

func TestWardenGet(t *testing.T) {
	r := gin.New()
	SetupRoutes(r)

	w := performRequest(r, "GET", "/allowed")
	assert.Equal(t, w.Code, http.StatusNotFound)
}

func TestWardenAnonymous(t *testing.T) {
	r := gin.New()
	SetupRoutes(r)

	w := performRequest(r, "POST", "/allowed")
	assert.Equal(t, w.Code, http.StatusUnauthorized)
}

func TestWardenWrongUsername(t *testing.T) {
	r := gin.New()
	SetupRoutes(r)

	req, _ := http.NewRequest("POST", "/allowed", nil)
	req.SetBasicAuth("alice", "chains")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusUnauthorized)
}

func performAllowed(t *testing.T, r *gin.Engine, body io.Reader, expected int, response interface{}) {
	req, _ := http.NewRequest("POST", "/allowed", body)
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("foo", "bar")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, expected, w.Code)
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.Nil(t, err)
}

func TestWardenEmpty(t *testing.T) {
	r := gin.New()
	SetupRoutes(r)

	var response ErrorResponse
	performAllowed(t, r, nil, http.StatusBadRequest, &response)
	assert.Equal(t, response.Message, "Missing body")
}

func TestWardenInvalidJSON(t *testing.T) {
	r := gin.New()
	SetupRoutes(r)

	body := bytes.NewBuffer([]byte("{\"random\\;mess\"}"))
	var response ErrorResponse
	performAllowed(t, r, body, http.StatusBadRequest, &response)
	assert.Contains(t, response.Message, "invalid character ';'")
}

func TestWardenAllowed(t *testing.T) {
	r := gin.New()
	SetupRoutes(r)

	for _, request := range []*ladon.Request{
		&ladon.Request{
			Subject:  "foo",
			Action:   "update",
			Resource: "server.org/blocklist:onecrl",
		},
		&ladon.Request{
			Subject:  "foo",
			Action:   "update",
			Resource: "server.org/blocklist:onecrl",
			Context: ladon.Context{
				"planet": "Mars",  // "mars" is case-sensitive
			},
		},
	} {
		token, _ := json.Marshal(request)
		body := bytes.NewBuffer(token)
		var response Response
		performAllowed(t, r, body, http.StatusOK, &response)
		assert.Equal(t, true, response.Allowed)
	}
}

func TestWardenNotAllowed(t *testing.T) {
	r := gin.New()
	SetupRoutes(r)

	for _, request := range []*ladon.Request{
		&ladon.Request{
			Subject:  "foo",
			Action:   "delete",
			Resource: "server.org/blocklist:onecrl",
		},
		&ladon.Request{
			Subject:  "foo",
			Action:   "update",
			Resource: "server.org/blocklist:onecrl",
			Context: ladon.Context{
				"planet": "mars",
			},
		},
	} {
		token, _ := json.Marshal(request)
		body := bytes.NewBuffer(token)
		var response Response
		performAllowed(t, r, body, http.StatusOK, &response)
		assert.Equal(t, false, response.Allowed)
	}
}
