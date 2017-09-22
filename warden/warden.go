package warden

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/ory/ladon"
	manager "github.com/ory/ladon/manager/memory"
	"gopkg.in/yaml.v2"

	"github.com/leplatrem/iam/utilities"
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

	var policies []*ladon.DefaultPolicy

	// Ladon does not support un/marshaling YAML.
	// XXX: I chose to convert to JSON first :|
	// https://github.com/ory/ladon/issues/83
	var generic interface{}
	if err := yaml.Unmarshal(yamlFile, &generic); err != nil {
		return nil, err
	}
	asJSON := utilities.Yaml2JSON(generic)
	jsonData, err := json.Marshal(asJSON)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(jsonData, &policies); err != nil {
		return nil, err
	}

	if len(policies) == 0 {
		log.Warning("No policies found.")
	}

	// []*ladon.DefaultPolicy does not implement ladon.Policy (missing AllowAccess method)
	// This casts as type ladon.Policies ([]ladon.Policy)
	result := make(ladon.Policies, len(policies))
	for i, pol := range policies {
		result[i] = pol
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
		policiesFile = filepath.Join(here, "policies.yaml")
	}
	policies, err := LoadPolicies(policiesFile)
	if err != nil {
		log.Fatal(err.Error())
	}

	for _, pol := range policies {
		log.Info("Load policy ", pol.GetID() + ": ", pol.GetDescription())
		err := warden.Manager.Create(pol)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

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
	allowed := (err == nil)

	// Show some debug information about matched policy.
	if allowed && gin.Mode() != gin.ReleaseMode {
		policies, _ := warden.Manager.FindRequestCandidates(&accessRequest)
		matched := policies[0]
		log.Debug("Policy matched ", matched.GetID() + ": ", matched.GetDescription())
	}

	c.JSON(http.StatusOK, gin.H{
		"allowed": allowed,
	})
}
