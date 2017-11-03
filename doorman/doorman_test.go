package doorman

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/ory/ladon"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	//Set Gin to Test Mode
	gin.SetMode(gin.TestMode)
	// Run the other tests
	os.Exit(m.Run())
}

func loadTempFiles(contents ...string) (*Doorman, error) {
	var filenames []string
	for _, content := range contents {
		tmpfile, _ := ioutil.TempFile("", "")
		defer os.Remove(tmpfile.Name()) // clean up
		tmpfile.Write([]byte(content))
		tmpfile.Close()
		filenames = append(filenames, tmpfile.Name())
	}
	return New(filenames, "")
}

func TestLoadBadPolicies(t *testing.T) {
	// Loads policies.yaml in current folder by default.
	_, err := New([]string{}, "")
	assert.NotNil(t, err) // doorman/policies.yaml does not exists.

	// Missing file
	_, err = New([]string{"/tmp/unknown.yaml"}, "")
	assert.NotNil(t, err)

	// Empty file
	_, err = loadTempFiles("")
	assert.NotNil(t, err)

	// Bad YAML
	_, err = loadTempFiles("$\\--xx")
	assert.NotNil(t, err)

	// Empty audience
	_, err = loadTempFiles(`
audience:
policies:
  -
    id: "1"
    effect: allow
`)
	assert.NotNil(t, err)

	// Empty policies
	_, err = loadTempFiles(`
audience: a
policies:
`)
	assert.Nil(t, err)

	// Bad audience
	_, err = loadTempFiles(`
audience: 1
policies:
  -
    id: "1"
    effect: allow
`)
	assert.NotNil(t, err)

	// Bad policies conditions
	_, err = loadTempFiles(`
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
	_, err = loadTempFiles(`
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
	_, err = loadTempFiles(`
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

func TestLoadGroups(t *testing.T) {
	d, err := loadTempFiles(`
audience: a
tags:
  admins:
    - alice@mit.edu
    - ldap|bob
  editors:
    - mathieu@mozilla.com
policies:
  -
    id: "1"
    effect: allow
`)
	assert.Nil(t, err)
	assert.Equal(t, len(d.tags["a"]), 2)
	assert.Equal(t, len(d.tags["a"]["admins"]), 2)
	assert.Equal(t, len(d.tags["a"]["editors"]), 1)
}

func TestReloadPolicies(t *testing.T) {
	doorman, err := New([]string{"../sample.yaml"}, "")
	assert.Nil(t, err)
	loaded, _ := doorman.ladons["https://sample.yaml"].Manager.GetAll(0, maxInt)
	assert.Equal(t, 6, len(loaded))

	// Second load.
	doorman.loadPolicies()
	loaded, _ = doorman.ladons["https://sample.yaml"].Manager.GetAll(0, maxInt)
	assert.Equal(t, 6, len(loaded))
}

func TestIsAllowed(t *testing.T) {
	doorman, err := New([]string{"../sample.yaml"}, "")
	assert.Nil(t, err)

	// Policy #1
	request := &Request{
		Principals: Principals{"userid:foo"},
		Action:     "update",
		Resource:   "server.org/blocklist:onecrl",
	}

	audience := "https://sample.yaml"

	// Check audience
	allowed, _ := doorman.IsAllowed(audience, request)
	assert.True(t, allowed)
	allowed, _ = doorman.IsAllowed("https://bad.audience", request)
	assert.False(t, allowed)

	// Expand principals from tags
	request = &Request{
		Principals: Principals{"userid:maria"},
	}
	_, principals := doorman.IsAllowed(audience, request)
	assert.Equal(t, principals, Principals{"userid:maria", "tag:admins"})

	// Expand principals from context roles
	request = &Request{
		Principals: Principals{"userid:bob"},
		Action: "update",
		Resource: "pto",
		Context: ladon.Context{
			"roles": []string{"editor"},
		},
	}
	_, principals = doorman.IsAllowed(audience, request)
	assert.Equal(t, principals, Principals{"userid:bob", "role:editor"})
}
