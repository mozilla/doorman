package warden

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ory/ladon"
	manager "github.com/ory/ladon/manager/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Response struct {
	Allowed bool
}

type ErrorResponse struct {
	Message string
}

const samplePoliciesFile string = "../sample.yaml"

var defaultConfig = Config{PoliciesFilename: "../policies.yaml"}

func TestMain(m *testing.M) {
	//Set Gin to Test Mode
	gin.SetMode(gin.TestMode)
	// Run the other tests
	os.Exit(m.Run())
}

func loadTempFile(warden *ladon.Ladon, content []byte) error {
	tmpfile, _ := ioutil.TempFile("", "")
	defer os.Remove(tmpfile.Name()) // clean up
	tmpfile.Write(content)
	tmpfile.Close()
	return LoadPolicies(warden, "bad.yaml")
}

func TestLoadPolicies(t *testing.T) {
	warden := &ladon.Ladon{
		Manager: manager.NewMemoryManager(),
	}

	// Missing file
	var err error
	err = LoadPolicies(warden, "/tmp/unknown.yaml")
	assert.NotNil(t, err)

	// Bad YAML
	err = loadTempFile(warden, []byte("$\\--xx"))
	assert.NotNil(t, err)

	// Bad policies
	err = loadTempFile(warden, []byte(`
	-
	  id: "1"
	  conditions:
	    - a
	    - b
	`))
	assert.NotNil(t, err)

	// Duplicated ID
	err = loadTempFile(warden, []byte(`
	-
	  id: "1"
	  effect: allow
	-
	  id: "1"
	  effect: deny
	`))
	assert.NotNil(t, err)
}

func performRequest(r http.Handler, method, path string, body io.Reader) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, body)
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

func TestWardenGet(t *testing.T) {
	r := gin.New()
	SetupRoutes(r, New(&defaultConfig))

	w := performRequest(r, "GET", "/allowed", nil)
	assert.Equal(t, w.Code, http.StatusNotFound)
}

func TestWardenEmpty(t *testing.T) {
	r := gin.New()
	SetupRoutes(r, New(&defaultConfig))

	var response ErrorResponse
	performAllowed(t, r, nil, http.StatusBadRequest, &response)
	assert.Equal(t, response.Message, "Missing body")
}

func TestWardenInvalidJSON(t *testing.T) {
	r := gin.New()
	SetupRoutes(r, New(&defaultConfig))

	body := bytes.NewBuffer([]byte("{\"random\\;mess\"}"))
	var response ErrorResponse
	performAllowed(t, r, body, http.StatusBadRequest, &response)
	assert.Contains(t, response.Message, "invalid character ';'")
}

func TestWardenAllowed(t *testing.T) {
	r := gin.New()
	warden := New(&defaultConfig)
	LoadPolicies(warden, samplePoliciesFile)
	SetupRoutes(r, warden)

	for _, request := range []*ladon.Request{
		// Policy #1
		{
			Subject:  "foo",
			Action:   "update",
			Resource: "server.org/blocklist:onecrl",
		},
		// Policy #2
		{
			Subject:  "foo",
			Action:   "update",
			Resource: "server.org/blocklist:onecrl",
			Context: ladon.Context{
				"planet": "Mars", // "mars" is case-sensitive
			},
		},
		// Policy #3
		{
			Subject:  "foo",
			Action:   "read",
			Resource: "server.org/blocklist:onecrl",
			Context: ladon.Context{
				"ip": "127.0.0.1",
			},
		},
		// Policy #4
		{
			Subject:  "bilbo",
			Action:   "wear",
			Resource: "ring",
			Context: ladon.Context{
				"owner": "bilbo",
			},
		},
		// Policy #5
		{
			Subject:  "group:admins",
			Action:   "create",
			Resource: "dns://",
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

func TestWardenNotAllowed(t *testing.T) {
	r := gin.New()
	warden := New(&defaultConfig)
	LoadPolicies(warden, samplePoliciesFile)
	SetupRoutes(r, warden)

	for _, request := range []*ladon.Request{
		// Policy #1
		{
			Subject:  "foo",
			Action:   "delete",
			Resource: "server.org/blocklist:onecrl",
		},
		// Policy #2
		{
			Subject:  "foo",
			Action:   "update",
			Resource: "server.org/blocklist:onecrl",
			Context: ladon.Context{
				"planet": "mars",
			},
		},
		// Policy #3
		{
			Subject:  "foo",
			Action:   "read",
			Resource: "server.org/blocklist:onecrl",
			Context: ladon.Context{
				"ip": "10.0.0.1",
			},
		},
		// Policy #4
		{
			Subject:  "gollum",
			Action:   "wear",
			Resource: "ring",
			Context: ladon.Context{
				"owner": "bilbo",
			},
		},
		// Policy #5
		{
			Subject:  "group:admins",
			Action:   "create",
			Resource: "dns://",
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
