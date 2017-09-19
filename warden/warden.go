package warden

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ory/ladon"
	manager "github.com/ory/ladon/manager/memory"
)

const ContextKey string = "warden"

type Token struct {
	Subject  string `form:"subject" json:"subject" binding:"required"`
	Action   string `form:"action" json:"action" binding:"required"`
	Resource string `form:"resource" json:"resource" binding:"required"`
}

func LadonMiddleware(warden *ladon.Ladon) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(ContextKey, warden)
		c.Next()
	}
}

// Warden views to query the policies.
func SetupRoutes(r *gin.Engine) {
	warden := &ladon.Ladon{
		Manager: manager.NewMemoryManager(),
	}

	// XXX: hard-coded policy
	warden.Manager.Create(&ladon.DefaultPolicy{
		ID:          "1",
		Description: "This policy allows foo to update any resource",
		Subjects:    []string{"foo"},
		Actions:     []string{"update"},
		Resources:   []string{"<.*>"},
		Effect:      ladon.AllowAccess,
	})

	// XXX: require Auth (currently hard-coded BasicAuth)
	authorized := r.Group("", gin.BasicAuth(gin.Accounts{
		"foo": "bar",
	}))

	authorized.Use(LadonMiddleware(warden))

	authorized.POST("/allowed", allowedHandler)
}

func allowedHandler(c *gin.Context) {
	if c.Request.ContentLength == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Missing body",
		})
		return
	}

	warden := c.MustGet(ContextKey).(*ladon.Ladon)

	var accessRequest ladon.Request
	if err := c.BindJSON(&accessRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	err := warden.IsAllowed(&accessRequest)

	c.JSON(http.StatusOK, gin.H{
		"allowed": (err == nil),
	})
}
