package doorman

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ory/ladon"
	log "github.com/sirupsen/logrus"
	jwt "gopkg.in/square/go-jose.v2/jwt"
)

// DoormanContextKey is the Gin context key to obtain the *Doorman instance.
const DoormanContextKey string = "doorman"

// ContextMiddleware adds the Doorman instance to the Gin context.
func ContextMiddleware(doorman *Doorman) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(DoormanContextKey, doorman)
		c.Next()
	}
}

// SetupRoutes adds doorman views to query the policies.
func SetupRoutes(r *gin.Engine, doorman *Doorman) {
	r.Use(ContextMiddleware(doorman))
	if doorman.JWTIssuer != "" {
		validator := &Auth0Validator{
			Issuer: doorman.JWTIssuer,
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

	var accessRequest ladon.Request
	if err := c.BindJSON(&accessRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	payloadJWT, ok := c.Get(JWTContextKey)
	// Is VerifyJWTMiddleware enabled? (disabled in tests)
	if ok {
		claims := payloadJWT.(*jwt.Claims)
		// Subject is taken from JWT.
		accessRequest.Subject = claims.Subject
	}

	doorman := c.MustGet(DoormanContextKey).(*Doorman)

	origin := c.Request.Header.Get("Origin")

	// Will fail if origin is unknown.
	err := doorman.IsAllowed(origin, &accessRequest)
	allowed := (err == nil)

	log.WithFields(
		log.Fields{
			"allowed":  allowed,
			"subject":  accessRequest.Subject,
			"action":   accessRequest.Action,
			"resource": accessRequest.Resource,
		},
	).Info("request.authorization")

	c.JSON(http.StatusOK, gin.H{
		"allowed": allowed,
		"user": gin.H{
			"id": accessRequest.Subject,
		},
	})
}
