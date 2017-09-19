package warden

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

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

func TestWardenPostEmpty(t *testing.T) {
    r := gin.New()
    SetupRoutes(r)

    w := performRequest(r, "POST", "/allowed")
    assert.Equal(t, w.Code, http.StatusOK)

    type Response struct {
        Allowed bool
    }
    var response Response
    err := json.Unmarshal(w.Body.Bytes(), &response)
    require.Nil(t, err)

    assert.False(t, response.Allowed)
}
