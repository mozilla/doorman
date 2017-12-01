package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mozilla/doorman/doorman"
)

func TestMain(m *testing.M) {
	AddLoader(&FileLoader{})
	AddLoader(&GithubLoader{
		Token: "",
	})

	// Run the other tests
	os.Exit(m.Run())
}

func loadTempFiles(contents ...string) (doorman.ServicesConfig, error) {
	var filenames []string
	for _, content := range contents {
		tmpfile, _ := ioutil.TempFile("", "")
		defer os.Remove(tmpfile.Name()) // clean up
		tmpfile.Write([]byte(content))
		tmpfile.Close()
		filenames = append(filenames, tmpfile.Name())
	}
	return Load(filenames)
}

func TestLoadBadPolicies(t *testing.T) {
	// Missing file
	_, err := Load([]string{"/tmp/unknown.yaml"})
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

	configs, err := Load([]string{dir})
	assert.Nil(t, err)
	require.Equal(t, len(configs), 1)
	assert.Equal(t, len(configs[0].Policies), 1)
}

func TestLoadGithub(t *testing.T) {
	// Unsupported URL
	_, err := Load([]string{"https://bitbucket.org/test.yaml"})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "no appropriate loader found")

	// Unsupported folder.
	_, err = Load([]string{"https://github.com/moz/ops/configs/"})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "not supported")

	// Bad URL
	_, err = Load([]string{"ftp://github.com/moz/ops/config.yaml"})
	assert.NotNil(t, err)

	// Bad file
	_, err = Load([]string{"https://github.com/mozilla/doorman/raw/06a2531/main.go"})
	assert.NotNil(t, err)

	// Good URL
	configs, err := Load([]string{"https://github.com/mozilla/doorman/raw/452ef7a/sample.yaml"})
	assert.Nil(t, err)
	require.Equal(t, len(configs), 1)
	assert.Equal(t, len(configs[0].Tags), 1)
	assert.Equal(t, len(configs[0].Policies), 6)
}

func TestLoadTags(t *testing.T) {
	configs, err := loadTempFiles(`
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
	require.Equal(t, len(configs), 1)
	assert.Equal(t, len(configs[0].Tags), 2)
	assert.Equal(t, len(configs[0].Tags["admins"]), 2)
	assert.Equal(t, len(configs[0].Tags["editors"]), 1)
}
