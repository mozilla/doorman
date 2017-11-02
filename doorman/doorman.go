package doorman

import (
	"encoding/json"
	"fmt"
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

// DoormanContextKey is the Gin context key to obtain the *Doorman instance.
const DoormanContextKey string = "doorman"

// JWTContextKey is the Gin context key to obtain the *jwt.Claims instance.
const JWTContextKey string = "JWT"

const maxInt int64 = 1<<63 - 1

// Doorman is the backend in charge of checking requests against policies.
type Doorman struct {
	PoliciesFilenames []string
	JWTIssuer         string
	ladons            map[string]ladon.Ladon
}

// Configuration represents the policies file content.
type Configuration struct {
	Audience string
	Policies []*ladon.DefaultPolicy
}

// New instantiates a new doorman.
func New(filenames []string, issuer string) (*Doorman, error) {
	// If not specified, read default file in current directory `./policies.yaml`
	if len(filenames) == 0 {
		here, _ := os.Getwd()
		filename := filepath.Join(here, DefaultPoliciesFilename)
		filenames = []string{filename}
	}

	w := &Doorman{filenames, issuer, map[string]ladon.Ladon{}}
	if err := w.loadPolicies(); err != nil {
		return nil, err
	}
	return w, nil
}

// IsAllowed is responsible for deciding if subject can perform action on a resource with a context.
func (doorman *Doorman) IsAllowed(audience string, request *ladon.Request) error {
	ladon, ok := doorman.ladons[audience]
	if !ok {
		return fmt.Errorf("unknown audience %q", audience)
	}
	return ladon.IsAllowed(request)
}

// LoadPolicies (re)loads configuration and policies from the YAML files.
func (doorman *Doorman) loadPolicies() error {
	// Clear every existing policy, and load new ones.
	for audience := range doorman.ladons {
		delete(doorman.ladons, audience)
	}
	// Load each configuration file.
	for _, filename := range doorman.PoliciesFilenames {
		log.Info("Load configuration ", filename)
		config, err := loadConfiguration(filename)
		if err != nil {
			return err
		}

		l := ladon.Ladon{
			Manager: manager.NewMemoryManager(),
		}
		for _, pol := range config.Policies {
			log.Info("Load policy ", pol.GetID()+": ", pol.GetDescription())
			err := l.Manager.Create(pol)
			if err != nil {
				return err
			}
		}
		_, exists := doorman.ladons[config.Audience]
		if exists {
			return fmt.Errorf("duplicated audience %q (filename %q)", config.Audience, filename)
		}
		doorman.ladons[config.Audience] = l
	}
	return nil
}

func loadConfiguration(filename string) (*Configuration, error) {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if len(yamlFile) == 0 {
		return nil, fmt.Errorf("empty file %q", filename)
	}
	// Ladon does not support un/marshaling YAML.
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

	var config Configuration
	if err := json.Unmarshal(jsonData, &config); err != nil {
		return nil, err
	}

	if config.Audience == "" {
		return nil, fmt.Errorf("empty audience in %q", filename)
	}

	if len(config.Policies) == 0 {
		log.Warningf("no policies found in %q", filename)
	}

	return &config, nil
}

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
