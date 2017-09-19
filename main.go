package main

import (
	"github.com/gin-gonic/gin"
	"go.mozilla.org/mozlogrus"

	"iam/utilities"
	"iam/warden"
)

func main() {
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

	r.Run() // listen and serve on 0.0.0.0:$PORT (:8080)
}
