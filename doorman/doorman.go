package doorman

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"errors"

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
	l                ladon.Ladon
	Manager          ladon.Manager
	PoliciesFilename string
	JWTIssuer        string
}

// Configuration represents the policies file content.
type Configuration struct {
	Audience string
	Policies []*ladon.DefaultPolicy
}

// New instantiates a new doorman.
func New(filename string, issuer string) (*Doorman, error) {
	l := ladon.Ladon{
		Manager: manager.NewMemoryManager(),
	}
	w := &Doorman{l, l.Manager, filename, issuer}
	if err := w.loadPolicies(); err != nil {
		return nil, err
	}
	return w, nil
}

// IsAllowed is responsible for deciding if subject can perform action on a resource with a context.
func (doorman *Doorman) IsAllowed(request *ladon.Request) error {
	return doorman.l.IsAllowed(request)
}

// LoadPolicies reads policies from the YAML file.
func (doorman *Doorman) loadPolicies() error {
	// If not specified, read it from ENV or read local `.policies.yaml`
	filename := doorman.PoliciesFilename
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

	// Ladon does not support un/marshaling YAML.
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

	var config Configuration
	if err := json.Unmarshal(jsonData, &config); err != nil {
		return err
	}

	if config.Audience == "" {
		return errors.New("Empty audience in configuration.")
	}

	if len(config.Policies) == 0 {
		log.Warning("No policies found.")
	}

	// Clear every existing policy, and load new ones.
	existing, err := doorman.Manager.GetAll(0, maxInt)
	if err != nil {
		return err
	}
	for _, pol := range existing {
		err := doorman.Manager.Delete(pol.GetID())
		if err != nil {
			return err
		}
	}
	for _, pol := range config.Policies {
		log.Info("Load policy ", pol.GetID()+": ", pol.GetDescription())
		err := doorman.Manager.Create(pol)
		if err != nil {
			return err
		}
	}

	return nil
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
	if ok {
		// With VerifyJWTMiddleware, subject is overriden by JWT.
		// (disabled for tests)
		accessRequest.Subject = payloadJWT.(*jwt.Claims).Subject
	}

	doorman := c.MustGet(DoormanContextKey).(*Doorman)
	err := doorman.IsAllowed(&accessRequest)
	allowed := (err == nil)

	// Show some information about matched policy.
	matchedInfo := gin.H{}
	if allowed {
		policies, _ := doorman.Manager.FindRequestCandidates(&accessRequest)
		matched := policies[0]
		matchedInfo = gin.H{
			"id":          matched.GetID(),
			"description": matched.GetDescription(),
		}
	}

	log.WithFields(
		log.Fields{
			"allowed":  allowed,
			"subject":  accessRequest.Subject,
			"action":   accessRequest.Action,
			"resource": accessRequest.Resource,
			"policy":   matchedInfo,
		},
	).Info("request.authorization")

	c.JSON(http.StatusOK, gin.H{
		"allowed": allowed,
		"policy":  matchedInfo,
		"user": gin.H{
			"id": accessRequest.Subject,
		},
	})
}
