package doorman

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// DoormanContextKey is the Gin context key to obtain the *Doorman instance.
const DoormanContextKey string = "doorman"

// ContextMiddleware adds the Doorman instance to the Gin context.
func ContextMiddleware(doorman Doorman) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(DoormanContextKey, doorman)
		c.Next()
	}
}

// SetupRoutes adds doorman views to query the policies.
func SetupRoutes(r *gin.Engine, doorman Doorman) {
	r.Use(ContextMiddleware(doorman))
	if jwtIssuer := doorman.JWTIssuer(); jwtIssuer != "" {
		// XXX: currently only Auth0 is supported.
		validator := &Auth0Validator{
			Issuer: jwtIssuer,
		}
		r.Use(VerifyJWTMiddleware(validator))
	} else {
		log.Warning("No JWT issuer configured. No authentication will be required.")
	}
	r.POST("/allowed", allowedHandler)
}

func allowedHandler(c *gin.Context) {
	if c.Request.ContentLength == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Missing body",
		})
		return
	}

	var accessRequest Request
	if err := c.BindJSON(&accessRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	doorman := c.MustGet(DoormanContextKey).(Doorman)
	audience := c.Request.Header.Get("Origin")

	// Is VerifyJWTMiddleware enabled?
	// If disabled (like in tests), principals can be posted in JSON.
	jwtPrincipals, ok := c.Get(PrincipalsContextKey)
	if ok {
		if len(accessRequest.Principals) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "cannot submit principals with JWT enabled",
			})
			return
		}
		accessRequest.Principals = jwtPrincipals.(Principals)
	}

	// Will fail if audience is unknown.
	allowed, principals := doorman.IsAllowed(audience, &accessRequest)

	authzLog.WithFields(
		log.Fields{
			"allowed":    allowed,
			"principals": principals,
			"action":     accessRequest.Action,
			"resource":   accessRequest.Resource,
		},
	).Info("request.authorization")

	c.JSON(http.StatusOK, gin.H{
		"allowed":    allowed,
		"principals": principals,
	})
}
