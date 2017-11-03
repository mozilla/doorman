package doorman

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ory/ladon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Response struct {
	Allowed    bool
	Principals Principals
}

type ErrorResponse struct {
	Message string
}

func performRequest(r http.Handler, method, path string, body io.Reader) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, body)
	req.Header.Set("Origin", "https://sample.yaml")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func performAllowed(t *testing.T, r *gin.Engine, body io.Reader, expected int, response interface{}) {
	w := performRequest(r, "POST", "/allowed", body)
	require.Equal(t, expected, w.Code)
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.Nil(t, err)
}

func TestDoormanGet(t *testing.T) {
	r := gin.New()
	doorman, _ := New([]string{"../sample.yaml"}, "")
	SetupRoutes(r, doorman)

	w := performRequest(r, "GET", "/allowed", nil)
	assert.Equal(t, w.Code, http.StatusNotFound)
}

func TestDoormanAllowed(t *testing.T) {
	r := gin.New()
	doorman, _ := New([]string{"../sample.yaml"}, "")
	SetupRoutes(r, doorman)

	for _, request := range []*Request{
		// Policy #1
		{
			Principals: []string{"userid:foo"},
			Action:     "update",
			Resource:   "server.org/blocklist:onecrl",
		},
		// Policy #2
		{
			Principals: []string{"userid:foo"},
			Action:     "update",
			Resource:   "server.org/blocklist:onecrl",
			Context: ladon.Context{
				"planet": "Mars", // "mars" is case-sensitive
			},
		},
		// Policy #3
		{
			Principals: []string{"userid:foo"},
			Action:     "read",
			Resource:   "server.org/blocklist:onecrl",
			Context: ladon.Context{
				"ip": "127.0.0.1",
			},
		},
		// Policy #4
		{
			Principals: []string{"userid:bilbo"},
			Action:     "wear",
			Resource:   "ring",
			Context: ladon.Context{
				"owner": "userid:bilbo",
			},
		},
		// Policy #5
		{
			Principals: []string{"group:admins"},
			Action:     "create",
			Resource:   "dns://",
			Context: ladon.Context{
				"domain": "kinto.mozilla.org",
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

func TestDoormanNotAllowed(t *testing.T) {
	r := gin.New()
	doorman, _ := New([]string{"../sample.yaml"}, "")
	SetupRoutes(r, doorman)

	for _, request := range []*Request{
		// Policy #1
		{
			Principals: []string{"userid:foo"},
			Action:     "delete",
			Resource:   "server.org/blocklist:onecrl",
		},
		// Policy #2
		{
			Principals: []string{"userid:foo"},
			Action:     "update",
			Resource:   "server.org/blocklist:onecrl",
			Context: ladon.Context{
				"planet": "mars",
			},
		},
		// Policy #3
		{
			Principals: []string{"userid:foo"},
			Action:     "read",
			Resource:   "server.org/blocklist:onecrl",
			Context: ladon.Context{
				"ip": "10.0.0.1",
			},
		},
		// Policy #4
		{
			Principals: []string{"userid:gollum"},
			Action:     "wear",
			Resource:   "ring",
			Context: ladon.Context{
				"owner": "bilbo",
			},
		},
		// Policy #5
		{
			Principals: []string{"group:admins"},
			Action:     "create",
			Resource:   "dns://",
			Context: ladon.Context{
				"domain": "kinto-storage.org",
			},
		},
		// Default
		{},
	} {
		token, _ := json.Marshal(request)
		body := bytes.NewBuffer(token)
		var response Response
		performAllowed(t, r, body, http.StatusOK, &response)
		assert.Equal(t, false, response.Allowed)
	}
}

func TestDoormanVerifiesJWT(t *testing.T) {
	r := gin.New()
	doorman, _ := New([]string{"../sample.yaml"}, "https://auth.mozilla.auth0.com/")
	SetupRoutes(r, doorman)

	// Policy #1 will match.
	authzRequest := Request{
		Principals: []string{"userid:foo"},
		Action:     "delete",
		Resource:   "server.org/blocklist:onecrl",
	}
	token, _ := json.Marshal(authzRequest)
	body := bytes.NewBuffer(token)
	var response ErrorResponse
	// Missing Authorization header.
	performAllowed(t, r, body, http.StatusUnauthorized, &response)
	assert.Equal(t, "Token not found", response.Message)
}

func TestAllowedHandlerBadRequest(t *testing.T) {
	var errResp ErrorResponse

	// Empty body
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request, _ = http.NewRequest("POST", "/allowed", nil)
	allowedHandler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.Equal(t, errResp.Message, "Missing body")

	// Invalid JSON
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)

	body := bytes.NewBuffer([]byte("{\"random\\;mess\"}"))
	c.Request, _ = http.NewRequest("POST", "/allowed", body)
	allowedHandler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.Contains(t, errResp.Message, "invalid character ';'")

	// Posted principals with JWT middleware enabled.
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	doorman, _ := New([]string{"../sample.yaml"}, "")
	c.Set(DoormanContextKey, doorman)
	c.Set(PrincipalsContextKey, Principals{"userid:maria"})  // Simulate JWT middleware.
	authzRequest := Request{
		Principals: Principals{"userid:superuser"},
	}
	post, _ := json.Marshal(authzRequest)
	body = bytes.NewBuffer(post)
	c.Request, _ = http.NewRequest("POST", "/allowed", body)
	allowedHandler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.Contains(t, errResp.Message, "cannot submit principals with JWT enabled")
}

func TestAllowedHandler(t *testing.T) {
	var resp Response

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	doorman, _ := New([]string{"../sample.yaml"}, "")
	c.Set(DoormanContextKey, doorman)

	// Using principals from context (JWT middleware)
	c.Set(PrincipalsContextKey, Principals{"userid:maria"})

	authzRequest := Request{
		Action:   "update",
	}
	post, _ := json.Marshal(authzRequest)
	body := bytes.NewBuffer(post)
	c.Request, _ = http.NewRequest("POST", "/allowed", body)
	c.Request.Header.Set("Origin", "https://sample.yaml")

	allowedHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.True(t, resp.Allowed)
	assert.Equal(t, Principals{"userid:maria", "tag:admins"}, resp.Principals)
}
