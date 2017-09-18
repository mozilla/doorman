package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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

func testJSONResponse(t *testing.T, url string, response interface{}) *httptest.ResponseRecorder {
	r := SetupRouter()

	req, _ := http.NewRequest("GET", url, nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fail()
	}

	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fail()
	}

	return w
}

func TestLBHeartbeat(t *testing.T) {
	type Response struct {
		Ok bool
	}
	var response Response
	testJSONResponse(t, "/__lbheartbeat__", &response)

	if !response.Ok {
		t.Fail()
	}
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

	if response.Commit != "stub" {
		t.Fail()
	}
}

func TestVersionMissing(t *testing.T) {
	os.Setenv("VERSION_FILE", "/tmp/missing.json")

	r := SetupRouter()
	req, _ := http.NewRequest("GET", "/__version__", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fail()
	}
}

func TestOpenAPI(t *testing.T) {
	type Response struct {
		Swagger string
	}
	var response Response
	testJSONResponse(t, "/__api__", &response)

	if response.Swagger != "2.0" {
		t.Fail()
	}
}

func TestContribute(t *testing.T) {
	type Response struct {
		Name string
	}
	var response Response
	testJSONResponse(t, "/contribute.json", &response)

	if response.Name != "IAM" {
		t.Fail()
	}
}
