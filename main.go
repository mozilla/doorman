package main

import (
	"os"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.mozilla.org/mozlogrus"

	"github.com/leplatrem/iam/utilities"
	"github.com/leplatrem/iam/warden"
)

func init() {
	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	default:
		if gin.Mode() == gin.ReleaseMode {
			log.SetLevel(log.InfoLevel)
		} else {
			log.SetLevel(log.DebugLevel)
		}
	}
}

func setupRouter() *gin.Engine {
	r := gin.New()
	// Crash free (turns errors into 5XX).
	r.Use(gin.Recovery())

	// Setup logging.
	if gin.Mode() == gin.ReleaseMode {
		// See https://github.com/mozilla-services/go-mozlogrus/issues/2#issuecomment-330495098
		r.Use(MozLogger())
		mozlogrus.Enable("iam")
	} else {
		r.Use(gin.Logger())
	}

	utilities.SetupRoutes(r)
	warden.SetupRoutes(r)

	return r
}

func main() {
	r := setupRouter()
	r.Run() // listen and serve on 0.0.0.0:$PORT (:8080)
}
