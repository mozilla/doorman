package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	//Set Gin to Test Mode
	gin.SetMode(gin.TestMode)
	// Run the other tests
	os.Exit(m.Run())
}

// Helper function to process a request and test its response
func testHTTPResponse(t *testing.T, r *gin.Engine, req *http.Request, f func(w *httptest.ResponseRecorder) bool) {
	// Create a response recorder
	w := httptest.NewRecorder()
	// Create the service and process the above request.
	r.ServeHTTP(w, req)
	if !f(w) {
		t.Fail()
	}
}

func TestLBHeartbeat(t *testing.T) {
	r := gin.Default()
	setupRoutes(r)

	req, _ := http.NewRequest("GET", "/__lbheartbeat__", nil)

	testHTTPResponse(t, r, req, func(w *httptest.ResponseRecorder) bool {
		statusOK := w.Code == http.StatusOK

		// Test that returned JSON contains "ok"
		p, err := ioutil.ReadAll(w.Body)
		jsonOK := err == nil && strings.Index(string(p), "\"ok\"") > 0

		return statusOK && jsonOK
	})
}

func TestHeartbeat(t *testing.T) {
	r := gin.Default()
	setupRoutes(r)

	req, _ := http.NewRequest("GET", "/__heartbeat__", nil)

	testHTTPResponse(t, r, req, func(w *httptest.ResponseRecorder) bool {
		statusOK := w.Code == http.StatusOK
		return statusOK
	})
}

func TestVersion(t *testing.T) {
	r := gin.Default()
	setupRoutes(r)

	req, _ := http.NewRequest("GET", "/__version__", nil)

	testHTTPResponse(t, r, req, func(w *httptest.ResponseRecorder) bool {
		statusOK := w.Code == http.StatusOK

		// Test that returned JSON contains "commit"
		p, err := ioutil.ReadAll(w.Body)
		jsonOK := err == nil && strings.Index(string(p), "\"commit\"") > 0

		return statusOK && jsonOK
	})
}

func TestVersionMissing(t *testing.T) {
	r := gin.Default()
	setupRoutes(r)

	os.Setenv("VERSION_FILE", "/tmp/missing.json")

	req, _ := http.NewRequest("GET", "/__version__", nil)

	testHTTPResponse(t, r, req, func(w *httptest.ResponseRecorder) bool {
		statusOK := w.Code == http.StatusNotFound
		return statusOK
	})
}

func TestOpenAPI(t *testing.T) {
	r := gin.Default()
	setupRoutes(r)

	req, _ := http.NewRequest("GET", "/__api__", nil)

	testHTTPResponse(t, r, req, func(w *httptest.ResponseRecorder) bool {
		statusOK := w.Code == http.StatusOK

		// Test that returned JSON contains "swagger"
		p, err := ioutil.ReadAll(w.Body)
		jsonOK := err == nil && strings.Index(string(p), "\"swagger\"") > 0

		return statusOK && jsonOK
	})
}
