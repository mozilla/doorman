package warden

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/ory/ladon"
	manager "github.com/ory/ladon/manager/memory"
	log "github.com/sirupsen/logrus"
	jwt "gopkg.in/square/go-jose.v2/jwt"
	"gopkg.in/yaml.v2"

	"github.com/leplatrem/iam/utilities"
)

// DefaultPoliciesFilename is the default policies filename.
const DefaultPoliciesFilename string = "policies.yaml"

// ContextKey is the Gin context key to obtain the *Warden instance.
const ContextKey string = "warden"

const maxInt int64 = 1<<63 - 1

// Config contains the settings of the warden.
type Config struct {
	PoliciesFilename string
	JWTIssuer        string
}

// Warden is the backend in charge of checking requests against policies.
type Warden struct {
	l       ladon.Ladon
	Manager ladon.Manager
	Config  *Config
}

// New instantiates a new warden.
func New(config *Config) *Warden {
	l := ladon.Ladon{
		Manager: manager.NewMemoryManager(),
	}
	w := &Warden{l, l.Manager, config}
	if err := w.LoadPolicies(config.PoliciesFilename); err != nil {
		log.Fatal(err.Error())
	}
	return w
}

// IsAllowed is responsible for deciding if subject can perform action on a resource with a context.
func (warden *Warden) IsAllowed(request *ladon.Request) error {
	return warden.l.IsAllowed(request)
}

// LoadPolicies reads policies from the YAML file.
func (warden *Warden) LoadPolicies(filename string) error {
	// If not specified, read it from ENV or read local `.policies.yaml`
	if filename == "" {
		filename = os.Getenv("POLICIES_FILE")
		if filename == "" {
			// Look in current working directory.
			here, _ := os.Getwd()
			filename = filepath.Join(here, DefaultPoliciesFilename)
		}
	}

	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	var policies []*ladon.DefaultPolicy

	// Ladon does not support un/marshaling YAML.
	// XXX: I chose to convert to JSON first :|
	// https://github.com/ory/ladon/issues/83
	var generic interface{}
	if err := yaml.Unmarshal(yamlFile, &generic); err != nil {
		return err
	}
	asJSON := utilities.Yaml2JSON(generic)
	jsonData, err := json.Marshal(asJSON)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(jsonData, &policies); err != nil {
		return err
	}

	if len(policies) == 0 {
		log.Warning("No policies found.")
	}

	// Clear every existing policy, and load new ones.
	existing, err := warden.Manager.GetAll(0, maxInt)
	if err != nil {
		return err
	}
	for _, pol := range existing {
		err := warden.Manager.Delete(pol.GetID())
		if err != nil {
			return err
		}
	}
	for _, pol := range policies {
		log.Info("Load policy ", pol.GetID()+": ", pol.GetDescription())
		err := warden.Manager.Create(pol)
		if err != nil {
			return err
		}
	}

	return nil
}

// ContextMiddleware adds the Warden instance to the Gin context.
func ContextMiddleware(warden *Warden) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(ContextKey, warden)
		c.Next()
	}
}

// SetupRoutes adds warden views to query the policies.
func SetupRoutes(r *gin.Engine, warden *Warden) {
	r.Use(ContextMiddleware(warden))
	if warden.Config.JWTIssuer != "" {
		r.Use(VerifyJWTMiddleware(warden.Config.JWTIssuer))
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

	payloadJWT, ok := c.Get("JWT")
	if ok {
		// With VerifyJWTMiddleware, subject is overriden by JWT.
		// (disabled for tests)
		accessRequest.Subject = payloadJWT.(*jwt.Claims).Subject
	}

	warden := c.MustGet(ContextKey).(*Warden)
	err := warden.IsAllowed(&accessRequest)
	allowed := (err == nil)

	// Show some debug information about matched policy.
	if allowed && gin.Mode() != gin.ReleaseMode {
		policies, _ := warden.Manager.FindRequestCandidates(&accessRequest)
		matched := policies[0]
		log.Debug("Policy matched ", matched.GetID()+": ", matched.GetDescription())
	}

	c.JSON(http.StatusOK, gin.H{
		"allowed": allowed,
	})
}
