package doorman

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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Policy struct {
	ID          string
	Description string
}

type User struct {
	ID string
}

type Response struct {
	Allowed bool
	User    User
	Policy  Policy
}

type ErrorResponse struct {
	Message string
}

const samplePoliciesFile string = "../sample.yaml"

var defaultConfig = Config{
	PoliciesFilename: "../policies.yaml",
	JWTIssuer:        "",
}

func TestMain(m *testing.M) {
	//Set Gin to Test Mode
	gin.SetMode(gin.TestMode)
	// Run the other tests
	os.Exit(m.Run())
}

func loadTempFile(doorman *Doorman, content []byte) error {
	tmpfile, _ := ioutil.TempFile("", "")
	defer os.Remove(tmpfile.Name()) // clean up
	tmpfile.Write(content)
	tmpfile.Close()
	return doorman.LoadPolicies(tmpfile.Name())
}

func TestLoadPolicies(t *testing.T) {
	var err error

	doorman := New(&Config{"../policies.yaml", ""})

	// Loads policies.yaml in current folder by default.
	err = doorman.LoadPolicies("")
	assert.NotNil(t, err) // doorman/policies.yaml does not exists.

	// Loads policies from env.
	os.Setenv("POLICIES_FILE", "/tmp/unknown.yaml")
	defer os.Unsetenv("POLICIES_FILE")
	err = doorman.LoadPolicies("")
	assert.NotNil(t, err)

	// Missing file
	err = doorman.LoadPolicies("/tmp/unknown.yaml")
	assert.NotNil(t, err)

	// Bad YAML
	err = loadTempFile(doorman, []byte("$\\--xx"))
	assert.NotNil(t, err)

	// Bad policies
	err = loadTempFile(doorman, []byte(`
	-
	  id: "1"
	  conditions:
	    - a
	    - b
	`))
	assert.NotNil(t, err)

	// Duplicated ID
	err = loadTempFile(doorman, []byte(`
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

func TestDoormanGet(t *testing.T) {
	r := gin.New()
	SetupRoutes(r, New(&defaultConfig))

	w := performRequest(r, "GET", "/allowed", nil)
	assert.Equal(t, w.Code, http.StatusNotFound)
}

func TestDoormanEmpty(t *testing.T) {
	r := gin.New()
	SetupRoutes(r, New(&defaultConfig))

	var response ErrorResponse
	performAllowed(t, r, nil, http.StatusBadRequest, &response)
	assert.Equal(t, response.Message, "Missing body")
}

func TestDoormanInvalidJSON(t *testing.T) {
	r := gin.New()
	SetupRoutes(r, New(&defaultConfig))

	body := bytes.NewBuffer([]byte("{\"random\\;mess\"}"))
	var response ErrorResponse
	performAllowed(t, r, body, http.StatusBadRequest, &response)
	assert.Contains(t, response.Message, "invalid character ';'")
}

func TestDoormanAllowed(t *testing.T) {
	r := gin.New()
	doorman := New(&defaultConfig)
	doorman.LoadPolicies(samplePoliciesFile)
	SetupRoutes(r, doorman)

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

func TestDoormanNotAllowed(t *testing.T) {
	r := gin.New()
	doorman := New(&defaultConfig)
	doorman.LoadPolicies(samplePoliciesFile)
	SetupRoutes(r, doorman)

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

func TestDoormanVerifiesJWT(t *testing.T) {
	r := gin.New()
	doorman := New(&Config{
		PoliciesFilename: "../policies.yaml",
		JWTIssuer:        "https://auth.mozilla.auth0.com/",
	})
	doorman.LoadPolicies("")
	SetupRoutes(r, doorman)

	// Policy #1 will match.
	request := ladon.Request{
		Subject:  "foo",
		Action:   "delete",
		Resource: "server.org/blocklist:onecrl",
	}
	token, _ := json.Marshal(request)
	body := bytes.NewBuffer(token)
	var response ErrorResponse

	// Missing Authorization header.
	performAllowed(t, r, body, http.StatusUnauthorized, &response)
	assert.Equal(t, "Token not found", response.Message)
}
