package config

import (
	"bytes"
	"os"
	"testing"

	"github.com/mozilla/doorman/doorman"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestLintingErrors(t *testing.T) {
	// Empty service
	c := doorman.ServiceConfig{
		Service: "",
	}
	err := lintConfigs(c)
	assert.NotNil(t, err)
}

func TestLintingWarnings(t *testing.T) {
	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	defer logrus.SetOutput(os.Stdout)

	// Empty policies
	c := doorman.ServiceConfig{
		Service: "abc",
	}
	err := lintConfigs(c)
	assert.Nil(t, err)
	assert.Contains(t, buf.String(), "No policies found in")
	buf.Reset()

	// HTTP verbs as actions
	c = doorman.ServiceConfig{
		Service: "abc",
		Policies: doorman.Policies{
			doorman.Policy{
				Actions: []string{"PUT"},
			},
		},
	}
	err = lintConfigs(c)
	assert.Nil(t, err)
	assert.Contains(t, buf.String(), "Avoid coupling of actions with HTTP verbs")
	buf.Reset()

	// HTTP verbs as actions
	c = doorman.ServiceConfig{
		Service: "abc",
		Policies: doorman.Policies{
			doorman.Policy{
				Actions:   []string{"read"},
				Resources: []string{"/articles/<.*>"},
			},
		},
	}
	err = lintConfigs(c)
	assert.Nil(t, err)
	assert.Contains(t, buf.String(), "Avoid coupling of resources with API URIs")
	buf.Reset()
}
