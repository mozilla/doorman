package doorman

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ory/ladon"
	"github.com/stretchr/testify/assert"
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

func TestMain(m *testing.M) {
	//Set Gin to Test Mode
	gin.SetMode(gin.TestMode)
	// Run the other tests
	os.Exit(m.Run())
}

func loadTempFiles(contents ...string) error {
	var filenames []string
	for _, content := range contents {
		tmpfile, _ := ioutil.TempFile("", "")
		defer os.Remove(tmpfile.Name()) // clean up
		tmpfile.Write([]byte(content))
		tmpfile.Close()
		filenames = append(filenames, tmpfile.Name())
	}
	_, err := New(filenames, "")
	return err
}

func TestLoadBadPolicies(t *testing.T) {
	// Loads policies.yaml in current folder by default.
	_, err := New([]string{}, "")
	assert.NotNil(t, err) // doorman/policies.yaml does not exists.

	// Missing file
	_, err = New([]string{"/tmp/unknown.yaml"}, "")
	assert.NotNil(t, err)

	// Empty file
	err = loadTempFiles("")
	assert.NotNil(t, err)

	// Bad YAML
	err = loadTempFiles("$\\--xx")
	assert.NotNil(t, err)

	// Empty audience
	err = loadTempFiles(`
audience:
policies:
  -
    id: "1"
    effect: allow
`)
	assert.NotNil(t, err)

	// Empty policies
	err = loadTempFiles(`
audience: a
policies:
`)
	assert.Nil(t, err)

	// Bad audience
	err = loadTempFiles(`
audience: 1
policies:
  -
    id: "1"
    effect: allow
`)
	assert.NotNil(t, err)

	// Bad policies conditions
	err = loadTempFiles(`
audience: a
policies:
  -
    id: "1"
    conditions:
      - a
      - b
`)
	assert.NotNil(t, err)

	// Duplicated policy ID
	err = loadTempFiles(`
audience: a
policies:
  -
    id: "1"
    effect: allow
  -
    id: "1"
    effect: deny
`)
	assert.NotNil(t, err)

	// Duplicated audience
	err = loadTempFiles(`
audience: a
policies:
  -
    id: "1"
    effect: allow
`, `
audience: a
policies:
  -
    id: "1"
    effect: allow
`)
	assert.NotNil(t, err)
}

func TestReloadPolicies(t *testing.T) {
	doorman, err := New([]string{"../sample.yaml"}, "")
	assert.Nil(t, err)
	loaded, _ := doorman.ladons["https://sample.yaml"].Manager.GetAll(0, maxInt)
	assert.Equal(t, 5, len(loaded))

	// Second load.
	doorman.loadPolicies()
	loaded, _ = doorman.ladons["https://sample.yaml"].Manager.GetAll(0, maxInt)
	assert.Equal(t, 5, len(loaded))
}

func TestIsAllowed(t *testing.T) {
	doorman, err := New([]string{"../sample.yaml"}, "")
	assert.Nil(t, err)

	request := &ladon.Request{
		// Policy #1
		Subject:  "foo",
		Action:   "update",
		Resource: "server.org/blocklist:onecrl",
	}
	assert.Nil(t, doorman.IsAllowed("https://sample.yaml", request))
	assert.NotNil(t, doorman.IsAllowed("https://bad.audience", request))
}
