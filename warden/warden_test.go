package warden

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

func TestWardenGet(t *testing.T) {
	r := gin.New()
	SetupRoutes(r)

	w := performRequest(r, "GET", "/allowed")
	assert.Equal(t, w.Code, http.StatusNotFound)
}

func TestWardenPostAnonymous(t *testing.T) {
	r := gin.New()
	SetupRoutes(r)

	w := performRequest(r, "POST", "/allowed")
	assert.Equal(t, w.Code, http.StatusUnauthorized)
}

func TestWardenPostEmpty(t *testing.T) {
    r := gin.New()
    SetupRoutes(r)

    req, _ := http.NewRequest("POST", "/allowed", nil)
    req.SetBasicAuth("foo", "bar")
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)

    assert.Equal(t, w.Code, http.StatusOK)

    type Response struct {
        Allowed bool
    }
    var response Response
    err := json.Unmarshal(w.Body.Bytes(), &response)
    require.Nil(t, err)

    assert.False(t, response.Allowed)
}
