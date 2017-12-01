package main

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// DefaultPoliciesFilename is the default policies filename.
const DefaultPoliciesFilename string = "policies.yaml"

var settings struct {
	GithubToken string
	Sources     []string
	LogLevel    logrus.Level
}

func sources() []string {
	// If POLICIES not specified, read ./policies.yaml
	env := os.Getenv("POLICIES")
	if env == "" {
		env = DefaultPoliciesFilename
	}
	sources := strings.Split(env, " ")
	// Filter empty strings
	var r []string
	for _, v := range sources {
		s := strings.TrimSpace(v)
		if s != "" {
			r = append(r, s)
		}
	}
	return r
}

func levelFromEnv() logrus.Level {
	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "fatal":
		return logrus.FatalLevel
	case "error":
		return logrus.ErrorLevel
	case "warn":
		return logrus.WarnLevel
	case "debug":
		return logrus.DebugLevel
	}
	// Default.
	if gin.Mode() == gin.ReleaseMode {
		return logrus.InfoLevel
	}
	return logrus.DebugLevel
}

func init() {
	settings.GithubToken = os.Getenv("GITHUB_TOKEN")
	settings.Sources = sources()
	settings.LogLevel = levelFromEnv()
}
