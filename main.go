package main

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/mozilla/doorman/doorman"
	"github.com/mozilla/doorman/utilities"
)

func setupRouter() (*gin.Engine, error) {
	r := gin.New()
	// Crash free (turns errors into 5XX).
	r.Use(gin.Recovery())

	// Setup logging.
	r.Use(HTTPLoggerMiddleware())

	// Setup doorman and load configuration files.
	w := doorman.NewDefaultLadon(doorman.Config{
		Sources: config.Sources,
	})
	if err := w.LoadPolicies(); err != nil {
		return nil, err
	}

	doorman.SetupRoutes(r, w)

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
