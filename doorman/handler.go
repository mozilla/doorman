package doorman

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupRoutes adds doorman views to query the policies.
func SetupRoutes(r *gin.Engine, doorman Doorman) {
	r.Use(ContextMiddleware(doorman))
	r.Use(VerifyJWTMiddleware(doorman))

	r.POST("/allowed", allowedHandler)
	r.POST("/__reload__", reloadHandler)
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

	// Is JWT verification enable for this audience?
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

	// Force some context values.
	if r.Context == nil {
		r.Context = Context{}
	}
	r.Context["audience"] = audience
	r.Context["remoteIP"] = c.Request.RemoteAddr

	// Expand principals with local ones.
	// Will do nothing if audience is unknown.
	r.Principals = doorman.ExpandPrincipals(audience, r.Principals)

	// Expand principals with specified roles.
	r.Principals = append(r.Principals, r.Roles()...)

	// Will deny if audience is unknown.
	allowed := doorman.IsAllowed(audience, &r)

	c.JSON(http.StatusOK, gin.H{
		"allowed":    allowed,
		"principals": r.Principals,
	})
}

func reloadHandler(c *gin.Context) {
	doorman := c.MustGet(DoormanContextKey).(Doorman)

	err := doorman.LoadPolicies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}
