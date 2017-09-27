package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/ory/ladon"
	"github.com/sirupsen/logrus"
	"go.mozilla.org/mozlogrus"

	"github.com/leplatrem/iam/utilities"
	"github.com/leplatrem/iam/warden"
)

func init() {
	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "fatal":
		logrus.SetLevel(logrus.FatalLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	default:
		if gin.Mode() == gin.ReleaseMode {
			logrus.SetLevel(logrus.InfoLevel)
		} else {
			logrus.SetLevel(logrus.DebugLevel)
		}
	}
}

func setupRouter(w *ladon.Ladon) *gin.Engine {
	// We disable mozlogrus for development.
	// See https://github.com/mozilla-services/go-mozlogrus/issues/2#issuecomment-330495098
	log.SetOutput(os.Stdout)

	r := gin.New()
	// Crash free (turns errors into 5XX).
	r.Use(gin.Recovery())

	// Setup logging.
	if gin.Mode() == gin.ReleaseMode {
		r.Use(MozLogger())
		mozlogrus.Enable("iam")
	} else {
		r.Use(gin.Logger())
	}

	utilities.SetupRoutes(r)
	warden.SetupRoutes(r, w)

	return r
}

func main() {
	// Load policies file from disk.
	w := warden.New()
	if err := warden.LoadPolicies(w, ""); err != nil {
		logrus.Fatal(err.Error())
	}

	r := setupRouter(w)
	r.Run() // listen and serve on 0.0.0.0:$PORT (:8080)
}
