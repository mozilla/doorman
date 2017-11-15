package main

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"github.com/leplatrem/iam/doorman"
	"github.com/leplatrem/iam/utilities"
)

func setupRouter() (*gin.Engine, error) {
	r := gin.New()
	// Crash free (turns errors into 5XX).
	r.Use(gin.Recovery())

	// Setup logging.
	r.Use(HTTPLoggerMiddleware())

	// Setup doorman and load configuration files.
	w := doorman.New(policies(), os.Getenv("JWT_ISSUER"))
	if err := w.LoadPolicies(); err != nil {
		return nil, err
	}

	doorman.SetupRoutes(r, w)

	utilities.SetupRoutes(r)

	return r, nil
}

func policies() []string {
	policies := strings.Split(os.Getenv("POLICIES"), " ")
	// Filter empty strings
	var r []string
	for _, v := range policies {
		s := strings.TrimSpace(v)
		if s != "" {
			r = append(r, s)
		}
	}
	return r
}

func main() {
	r, err := setupRouter()
	if err != nil {
		log.Fatal(err.Error())
	}
	r.Run() // listen and serve on 0.0.0.0:$PORT (:8080)
}
