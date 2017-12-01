package doorman

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
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

func sampleDoorman() *LadonDoorman {
	doorman := NewDefaultLadon(Config{
		Sources: []string{"../sample.yaml"},
	})
	doorman.LoadPolicies()
	return doorman
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
	w := NewDefaultLadon(Config{Sources: filenames})
	err := w.LoadPolicies()
	return w, err
}

func TestNewDefaultLadon(t *testing.T) {
	w := NewDefaultLadon(Config{
		Sources: []string{"some-file.yaml"},
	})
	assert.Equal(t, w.config.Sources[0], "some-file.yaml")
}

func TestLoadBadPolicies(t *testing.T) {
	// Missing file
	w := NewDefaultLadon(Config{Sources: []string{"/tmp/unknown.yaml"}})
	err := w.LoadPolicies()
	assert.NotNil(t, err)

	// Empty file
	_, err = loadTempFiles("")
	assert.NotNil(t, err)

	// Bad YAML
	_, err = loadTempFiles("$\\--xx")
	assert.NotNil(t, err)

	// Empty service
	_, err = loadTempFiles(`
service:
policies:
  -
    id: "1"
    effect: allow
`)
	assert.NotNil(t, err)

	// Empty policies
	_, err = loadTempFiles(`
service: a
policies:
`)
	assert.Nil(t, err)

	// Bad policies conditions
	_, err = loadTempFiles(`
service: a
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
service: a
policies:
  -
    id: "1"
    effect: allow
  -
    id: "1"
    effect: deny
`)
	assert.NotNil(t, err)

	// Duplicated service
	_, err = loadTempFiles(`
service: a
policies:
  -
    id: "1"
    effect: allow
`, `
service: a
policies:
  -
    id: "1"
    effect: allow
`)
	assert.NotNil(t, err)

	// Bad JWT issuer
	_, err = loadTempFiles(`
service: a
jwtIssuer: https://perlin-pinpin
policies:
  -
    id: "1"
    effect: allow
`)
	assert.NotNil(t, err)
}

func TestLoadFolder(t *testing.T) {
	// Create temp dir
	dir, err := ioutil.TempDir("", "example")
	assert.Nil(t, err)
	defer os.RemoveAll(dir)
	// Create subdir (to be skipped)
	subdir, err := ioutil.TempDir(dir, "ignored")
	assert.Nil(t, err)
	defer os.RemoveAll(subdir)

	// Create sample file
	testfile := filepath.Join(dir, "test.yaml")
	defer os.Remove(testfile)
	err = ioutil.WriteFile(testfile, []byte(`
service: a
policies:
  -
    id: "1"
    action: read
    effect: allow
`), 0666)

	w := NewDefaultLadon(Config{Sources: []string{dir}})
	err = w.LoadPolicies()
	assert.Nil(t, err)
	assert.Equal(t, len(w.services["a"].Policies), 1)
}

func TestLoadGithub(t *testing.T) {
	// Unsupported URL
	w := NewDefaultLadon(Config{Sources: []string{"https://bitbucket.org/test.yaml"}})
	err := w.LoadPolicies()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "no appropriate loader found")

	// Unsupported folder.
	w = NewDefaultLadon(Config{Sources: []string{"https://github.com/moz/ops/configs/"}})
	err = w.LoadPolicies()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "not supported")

	// Bad URL
	w = NewDefaultLadon(Config{Sources: []string{"ftp://github.com/moz/ops/config.yaml"}})
	err = w.LoadPolicies()
	assert.NotNil(t, err)

	// Bad file
	w = NewDefaultLadon(Config{Sources: []string{"https://github.com/mozilla/doorman/raw/06a2531/main.go"}})
	err = w.LoadPolicies()
	assert.NotNil(t, err)

	// Good URL
	w = NewDefaultLadon(Config{Sources: []string{"https://github.com/mozilla/doorman/raw/452ef7a/sample.yaml"}})
	err = w.LoadPolicies()
	assert.Nil(t, err)
	assert.Equal(t, len(w.services["https://sample.yaml"].Tags), 1)
	assert.Equal(t, len(w.services["https://sample.yaml"].Policies), 6)
}

func TestLoadTags(t *testing.T) {
	d, err := loadTempFiles(`
service: a
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
	assert.Equal(t, len(d.services["a"].Tags), 2)
	assert.Equal(t, len(d.services["a"].Tags["admins"]), 2)
	assert.Equal(t, len(d.services["a"].Tags["editors"]), 1)
}

func TestReloadPolicies(t *testing.T) {
	doorman := sampleDoorman()
	loaded, _ := doorman.ladons["https://sample.yaml"].Manager.GetAll(0, maxInt)
	assert.Equal(t, 6, len(loaded))

	// Second load.
	doorman.LoadPolicies()
	loaded, _ = doorman.ladons["https://sample.yaml"].Manager.GetAll(0, maxInt)
	assert.Equal(t, 6, len(loaded))

	// Load bad policies, does not affect existing.
	doorman.config.Sources = []string{"/tmp/unknown.yaml"}
	doorman.LoadPolicies()
	_, ok := doorman.ladons["https://sample.yaml"]
	assert.True(t, ok)
}

func TestIsAllowed(t *testing.T) {
	doorman := sampleDoorman()

	// Policy #1
	request := &Request{
		Principals: Principals{"userid:foo"},
		Action:     "update",
		Resource:   "server.org/blocklist:onecrl",
	}

	// Check service
	allowed := doorman.IsAllowed("https://sample.yaml", request)
	assert.True(t, allowed)
	allowed = doorman.IsAllowed("https://bad.service", request)
	assert.False(t, allowed)
}

func TestExpandPrincipals(t *testing.T) {
	doorman := sampleDoorman()

	// Expand principals from tags
	principals := doorman.ExpandPrincipals("https://sample.yaml", Principals{"userid:maria"})
	assert.Equal(t, principals, Principals{"userid:maria", "tag:admins"})
}

func TestDoormanAllowed(t *testing.T) {
	doorman := sampleDoorman()

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
		// Policy #4 (list of principals)
		{
			Principals: []string{"userid:bilbo"},
			Action:     "wear",
			Resource:   "ring",
			Context: Context{
				"owner": []string{"userid:alice", "userid:bilbo"},
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
	doorman := sampleDoorman()

	for _, request := range []*Request{
		// Policy #1
		{
			Principals: []string{"userid:foo"},
			Action:     "delete",
			Resource:   "server.org/blocklist:onecrl",
			Context:    Context{},
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
		{
			Context: Context{},
		},
	} {
		// Force context value like in handler.
		request.Context["_principals"] = request.Principals
		assert.Equal(t, false, doorman.IsAllowed("https://sample.yaml", request))
	}
}

func TestDoormanAuditLogger(t *testing.T) {
	doorman := sampleDoorman()

	var buf bytes.Buffer
	doorman.auditLogger().logger.Out = &buf
	defer func() {
		doorman.auditLogger().logger.Out = os.Stdout
	}()

	// Logs when service is bad.
	doorman.IsAllowed("bad service", &Request{})
	assert.Contains(t, buf.String(), "\"allowed\":false")

	service := "https://sample.yaml"

	// Logs policies.
	buf.Reset()
	doorman.IsAllowed(service, &Request{
		Principals: Principals{"userid:any"},
		Action:     "any",
		Resource:   "any",
		Context: Context{
			"planet":      "mars",
			"_principals": Principals{"userid:any"},
		},
	})
	assert.Contains(t, buf.String(), "\"allowed\":false")
	assert.Contains(t, buf.String(), "\"policies\":[\"2\"]")

	buf.Reset()
	doorman.IsAllowed(service, &Request{
		Principals: Principals{"userid:foo"},
		Action:     "update",
		Resource:   "server.org/blocklist:onecrl",
	})
	assert.Contains(t, buf.String(), "\"allowed\":true")
	assert.Contains(t, buf.String(), "\"policies\":[\"1\"]")
}
