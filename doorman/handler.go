package doorman

import (
	"fmt"
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

	var r Request
	if err := c.BindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	// Is VerifyJWTMiddleware enabled?
	// If disabled (like in tests), principals can be posted in JSON.
	jwtPrincipals, ok := c.Get(PrincipalsContextKey)
	if ok {
		if len(r.Principals) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "cannot submit principals with JWT enabled",
			})
			return
		}
		r.Principals = jwtPrincipals.(Principals)
	} else {
		if len(r.Principals) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "missing principals",
			})
			return
		}
	}

	doorman := c.MustGet(DoormanContextKey).(Doorman)
	audience := c.Request.Header.Get("Origin")

	// Expand principals with local ones.
	// Will do nothing if audience is unknown.
	r.Principals = doorman.ExpandPrincipals(audience, r.Principals)

	// Expand principals with specified roles.
	if roles, ok := r.Context["roles"]; ok {
		if rolesI, ok := roles.([]interface{}); ok {
			for _, roleI := range rolesI {
				if role, ok := roleI.(string); ok {
					prefixed := fmt.Sprintf("role:%s", role)
					r.Principals = append(r.Principals, prefixed)
				}
			}
		}
	}

	// Will deny if audience is unknown.
	allowed := doorman.IsAllowed(audience, &r)

	c.JSON(http.StatusOK, gin.H{
		"allowed":    allowed,
		"principals": r.Principals,
	})
}
