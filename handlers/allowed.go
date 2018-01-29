package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mozilla/doorman/doorman"
)

func allowedHandler(c *gin.Context) {
	if c.Request.ContentLength == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Missing body",
		})
		return
	}

	var r doorman.Request
	if err := c.BindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	// Is authentication verification enable for this service?
	// If disabled (like in tests), principals can be posted in JSON.
	principals, ok := c.Get(PrincipalsContextKey)
	if ok {
		if len(r.Principals) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "cannot submit principals with authentication enabled",
			})
			return
		}
		r.Principals = principals.(doorman.Principals)
	} else {
		if len(r.Principals) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "missing principals",
			})
			return
		}
	}

	d := c.MustGet(DoormanContextKey).(doorman.Doorman)
	service := c.Request.Header.Get("Origin")

	// Expand principals with local ones.
	r.Principals = d.ExpandPrincipals(service, r.Principals)
	// Expand principals with specified roles.
	r.Principals = append(r.Principals, r.Roles()...)

	// Force some context values (for Audit logger mainly)
	// XXX: using the context field to pass custom values on *ladon.Request
	// for audit logging is not very elegant.
	if r.Context == nil {
		r.Context = doorman.Context{}
	}
	r.Context["remoteIP"] = c.Request.RemoteAddr
	r.Context["_service"] = service
	r.Context["_principals"] = r.Principals

	allowed := d.IsAllowed(service, &r)

	c.JSON(http.StatusOK, gin.H{
		"allowed":    allowed,
		"principals": r.Principals,
	})
}
