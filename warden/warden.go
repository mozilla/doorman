package warden

import (
	"path/filepath"
	"net/http"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/ory/ladon"
	manager "github.com/ory/ladon/manager/memory"
	"gopkg.in/yaml.v2"
)

// ContextKey is the Gin context key to obtain the *ladon.Ladon instance.
const ContextKey string = "warden"

// LadonMiddleware adds the ladon.Ladon instance to the Gin context.
func LadonMiddleware(warden *ladon.Ladon) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(ContextKey, warden)
		c.Next()
	}
}

// LoadPolicies reads policies from the YAML file.
func LoadPolicies(filename string) (ladon.Policies, error) {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var policies []ladon.DefaultPolicy;
	if err := yaml.Unmarshal(yamlFile, &policies); err != nil {
		return nil, err
	}

	if len(policies) == 0 {
		log.Warning("No policies found.")
	}

	// XXX: []*ladon.DefaultPolicy does not implement ladon.Policy (missing AllowAccess method)
	result := make(ladon.Policies, len(policies))
	for i, pol := range policies {
		result[i] = &pol;
	}
	return result, nil
}

// SetupRoutes adds warden views to query the policies.
func SetupRoutes(r *gin.Engine) {
	warden := &ladon.Ladon{
		Manager: manager.NewMemoryManager(),
	}

	policiesFile := os.Getenv("POLICIES_FILE")
	if policiesFile == "" {
		// Look in current working directory.
		here, _ := os.Getwd()
		policiesFile = filepath.Join(here, "policies.yml")
	}
	policies, err := LoadPolicies(policiesFile)
	if err != nil {
		log.Fatal(err.Error())
	}

	for _, pol := range policies {
		err := warden.Manager.Create(pol)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	log.WithFields(log.Fields{
		"filename": policiesFile,
		"size": len(policies),
	}).Info("Policies file loaded.")

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
