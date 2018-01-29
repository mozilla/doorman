package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mozilla/doorman/config"
	"github.com/mozilla/doorman/doorman"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type AllowedResponse struct {
	Allowed    bool
	Principals doorman.Principals
}

type ErrorResponse struct {
	Message string
}

func performAllowed(t *testing.T, r *gin.Engine, body io.Reader, expected int, response interface{}) {
	w := performRequest(r, "POST", "/allowed", body)
	require.Equal(t, expected, w.Code)
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.Nil(t, err)
}

func TestAllowedGet(t *testing.T) {
	r := gin.New()
	d := doorman.NewDefaultLadon()
	SetupRoutes(r, d)

	w := performRequest(r, "GET", "/allowed", nil)
	assert.Equal(t, w.Code, http.StatusNotFound)
}

func TestAllowedVerifiesAuthentication(t *testing.T) {
	d := doorman.NewDefaultLadon()
	// Will initialize an authenticator (ie. download public keys)
	d.LoadPolicies(doorman.ServicesConfig{
		doorman.ServiceConfig{
			Service:   "https://sample.yaml",
			JWTIssuer: "https://auth.mozilla.auth0.com/",
			Policies: doorman.Policies{
				doorman.Policy{
					Actions: []string{"update"},
				},
			},
		},
	})

	r := gin.New()
	SetupRoutes(r, d)

	authzRequest := doorman.Request{}
	token, _ := json.Marshal(authzRequest)
	body := bytes.NewBuffer(token)
	var response ErrorResponse
	// Missing Authorization header.
	performAllowed(t, r, body, http.StatusUnauthorized, &response)
	assert.Equal(t, "token not found", response.Message)
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

	// Missing principals when AuthnMiddleware not enabled.
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)

	body = bytes.NewBuffer([]byte("{\"action\":\"update\"}"))
	c.Request, _ = http.NewRequest("POST", "/allowed", body)
	allowedHandler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.Contains(t, errResp.Message, "missing principals")

	d := doorman.NewDefaultLadon()

	// Posted principals with AuthnMiddleware enabled.
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Set(DoormanContextKey, d)
	c.Set(PrincipalsContextKey, doorman.Principals{"userid:maria"}) // Simulate authn middleware.
	authzRequest := doorman.Request{
		Principals: doorman.Principals{"userid:superuser"},
	}
	post, _ := json.Marshal(authzRequest)
	body = bytes.NewBuffer(post)
	c.Request, _ = http.NewRequest("POST", "/allowed", body)
	allowedHandler(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	json.Unmarshal(w.Body.Bytes(), &errResp)
	assert.Contains(t, errResp.Message, "cannot submit principals with authentication enabled")
}

func TestAllowedHandler(t *testing.T) {
	var resp AllowedResponse

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	configs, err := config.Load([]string{"../sample.yaml"})
	require.Nil(t, err)

	d := doorman.NewDefaultLadon()
	err = d.LoadPolicies(configs)
	require.Nil(t, err)
	c.Set(DoormanContextKey, d)

	// Using principals from context (AuthnMiddleware)
	c.Set(PrincipalsContextKey, doorman.Principals{"userid:maria"})

	authzRequest := doorman.Request{
		Action: "update",
	}
	post, _ := json.Marshal(authzRequest)
	body := bytes.NewBuffer(post)
	c.Request, _ = http.NewRequest("POST", "/allowed", body)
	c.Request.Header.Set("Origin", "https://sample.yaml")

	allowedHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.True(t, resp.Allowed)
	assert.Equal(t, doorman.Principals{"userid:maria", "tag:admins"}, resp.Principals)
}

func TestAllowedHandlerRoles(t *testing.T) {
	var resp AllowedResponse

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	configs, err := config.Load([]string{"../sample.yaml"})
	require.Nil(t, err)

	println(len(configs))
	d := doorman.NewDefaultLadon()
	err = d.LoadPolicies(configs)
	require.Nil(t, err)
	c.Set(DoormanContextKey, d)

	// Expand principals from context roles
	authzRequest := doorman.Request{
		Principals: doorman.Principals{"userid:bob"},
		Action:     "update",
		Resource:   "pto",
		Context: doorman.Context{
			"roles": []string{"editor"},
		},
	}
	post, _ := json.Marshal(authzRequest)
	body := bytes.NewBuffer(post)
	c.Request, _ = http.NewRequest("POST", "/allowed", body)

	allowedHandler(c)

	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, doorman.Principals{"userid:bob", "role:editor"}, resp.Principals)
}
