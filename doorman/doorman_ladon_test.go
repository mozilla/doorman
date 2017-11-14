package doorman

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	//Set Gin to Test Mode
	gin.SetMode(gin.TestMode)
	// Run the other tests
	os.Exit(m.Run())
}

func loadTempFiles(contents ...string) (*LadonDoorman, error) {
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

	// Check audience
	allowed := doorman.IsAllowed("https://sample.yaml", request)
	assert.True(t, allowed)
	allowed = doorman.IsAllowed("https://bad.audience", request)
	assert.False(t, allowed)
}

func TestExpandPrincipals(t *testing.T) {
	doorman, err := New([]string{"../sample.yaml"}, "")
	assert.Nil(t, err)

	// Expand principals from tags
	principals := doorman.ExpandPrincipals("https://sample.yaml", Principals{"userid:maria"})
	assert.Equal(t, principals, Principals{"userid:maria", "tag:admins"})
}

func TestDoormanAllowed(t *testing.T) {
	doorman, _ := New([]string{"../sample.yaml"}, "")

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
			Context: Context{
				"planet": "Mars", // "mars" is case-sensitive
			},
		},
		// Policy #3
		{
			Principals: []string{"userid:foo"},
			Action:     "read",
			Resource:   "server.org/blocklist:onecrl",
			Context: Context{
				"ip": "127.0.0.1",
			},
		},
		// Policy #4
		{
			Principals: []string{"userid:bilbo"},
			Action:     "wear",
			Resource:   "ring",
			Context: Context{
				"owner": "userid:bilbo",
			},
		},
		// Policy #5
		{
			Principals: []string{"group:admins"},
			Action:     "create",
			Resource:   "dns://",
			Context: Context{
				"domain": "kinto.mozilla.org",
			},
		},
	} {
		assert.Equal(t, true, doorman.IsAllowed("https://sample.yaml", request))
	}
}

func TestDoormanNotAllowed(t *testing.T) {
	doorman, _ := New([]string{"../sample.yaml"}, "")

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
			Context: Context{
				"planet": "mars",
			},
		},
		// Policy #3
		{
			Principals: []string{"userid:foo"},
			Action:     "read",
			Resource:   "server.org/blocklist:onecrl",
			Context: Context{
				"ip": "10.0.0.1",
			},
		},
		// Policy #4
		{
			Principals: []string{"userid:gollum"},
			Action:     "wear",
			Resource:   "ring",
			Context: Context{
				"owner": "bilbo",
			},
		},
		// Policy #5
		{
			Principals: []string{"group:admins"},
			Action:     "create",
			Resource:   "dns://",
			Context: Context{
				"domain": "kinto-storage.org",
			},
		},
		// Default
		{},
	} {
		assert.Equal(t, false, doorman.IsAllowed("https://sample.yaml", request))
	}
}

func TestDoormanAuditLogger(t *testing.T) {
	doorman, _ := New([]string{"../sample.yaml"}, "")

	var buf bytes.Buffer
	doorman.auditLogger().logger.Out = &buf
	defer func() {
		doorman.auditLogger().logger.Out = os.Stdout
	}()

	// Logs when audience is bad.
	doorman.IsAllowed("bad audience", &Request{})
	assert.Contains(t, buf.String(), "\"allowed\":false")

	audience := "https://sample.yaml"

	// Logs policies.
	buf.Reset()
	doorman.IsAllowed(audience, &Request{
		Principals: []string{"userid:any"},
		Action:     "any",
		Resource:   "any",
		Context: Context{
			"planet": "mars",
		},
	})
	assert.Contains(t, buf.String(), "\"allowed\":false")
	assert.Contains(t, buf.String(), "\"policies\":[\"2\"]")

	buf.Reset()
	doorman.IsAllowed(audience, &Request{
		Principals: []string{"userid:foo"},
		Action:     "update",
		Resource:   "server.org/blocklist:onecrl",
	})
	assert.Contains(t, buf.String(), "\"allowed\":true")
	assert.Contains(t, buf.String(), "\"policies\":[\"1\"]")
}
