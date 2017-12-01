package main

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/mozilla/doorman/config"
	"github.com/mozilla/doorman/doorman"
	"github.com/mozilla/doorman/utilities"
)

func init() {
	config.AddLoader(&config.FileLoader{})
	config.AddLoader(&config.GithubLoader{
		Token: settings.GithubToken,
	})
}

func setupRouter() (*gin.Engine, error) {
	r := gin.New()
	// Crash free (turns errors into 5XX).
	r.Use(gin.Recovery())

	// Setup logging.
	setupLogging()

	r.Use(HTTPLoggerMiddleware())

	// Setup doorman and load configuration files.
	d := doorman.NewDefaultLadon()

	// Load files (from folders, files, Github, etc.)
	configs, err := config.Load(settings.Sources)
	if err != nil {
		return nil, err
	}

	// Load into Doorman.
	if err := d.LoadPolicies(configs); err != nil {
		return nil, err
	}

	doorman.SetupRoutes(r, d)

	config.SetupRoutes(r, settings.Sources)

	utilities.SetupRoutes(r)

	return r, nil
}

func main() {
	r, err := setupRouter()
	if err != nil {
		log.Fatal(err.Error())
	}
	r.Run() // listen and serve on 0.0.0.0:$PORT (:8080)
}
