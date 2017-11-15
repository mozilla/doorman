package doorman

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ory/ladon"
	manager "github.com/ory/ladon/manager/memory"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/leplatrem/iam/utilities"
)

// DefaultPoliciesFilename is the default policies filename.
const DefaultPoliciesFilename string = "policies.yaml"

const maxInt int64 = 1<<63 - 1

// Tags map tag names to principals.
type Tags map[string]Principals

// LadonDoorman is the backend in charge of checking requests against policies.
type LadonDoorman struct {
	policiesSources []string
	jwtIssuer       string
	ladons          map[string]ladon.Ladon
	tags            map[string]Tags
	_auditLogger    *auditLogger
}

// Configuration represents the policies file content.
type Configuration struct {
	Audience string
	Tags     Tags
	Policies []*ladon.DefaultPolicy
}

// New instantiates a new doorman.
func New(policies []string, issuer string) *LadonDoorman {
	// If not specified, read default file in current directory `./policies.yaml`
	if len(policies) == 0 {
		here, _ := os.Getwd()
		filename := filepath.Join(here, DefaultPoliciesFilename)
		policies = []string{filename}
	}

	w := &LadonDoorman{
		policiesSources: policies,
		jwtIssuer:       issuer,
		ladons:          map[string]ladon.Ladon{},
		tags:            map[string]Tags{},
	}
	return w
}

func (doorman *LadonDoorman) auditLogger() *auditLogger {
	if doorman._auditLogger == nil {
		doorman._auditLogger = newAuditLogger()
	}
	return doorman._auditLogger
}

// LoadPolicies (re)loads configuration and policies from the YAML files.
func (doorman *LadonDoorman) LoadPolicies() error {
	// Clear every existing policy, and load new ones.
	for audience := range doorman.ladons {
		delete(doorman.ladons, audience)
	}

	// Load each configuration file.
	for _, filename := range doorman.policiesSources {
		log.Info("Load configuration ", filename)
		config, err := loadConfiguration(filename)
		if err != nil {
			return err
		}

		l := ladon.Ladon{
			Manager:     manager.NewMemoryManager(),
			AuditLogger: doorman.auditLogger(),
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
		doorman.tags[config.Audience] = config.Tags
	}
	return nil
}

// JWTIssuer returns the URL of the JWT issuer (if configured)
func (doorman *LadonDoorman) JWTIssuer() string {
	return doorman.jwtIssuer
}

// IsAllowed is responsible for deciding if subject can perform action on a resource with a context.
func (doorman *LadonDoorman) IsAllowed(audience string, request *Request) bool {
	// Instantiate objects from the ladon API.
	context := ladon.Context{}
	for key, value := range request.Context {
		context[key] = value
	}

	r := &ladon.Request{
		Resource: request.Resource,
		Action:   request.Action,
		Context:  context,
	}

	l, ok := doorman.ladons[audience]
	if !ok {
		// Explicitly log denied request using audit logger.
		doorman.auditLogger().logRequest(false, r, ladon.Policies{})
		return false
	}

	// For each principal, use it as the subject and query ladon backend.
	for _, principal := range request.Principals {
		r.Subject = principal
		if err := l.IsAllowed(r); err == nil {
			return true
		}
	}
	return false
}

// ExpandPrincipals will match the tags defined in the configuration for this audience
// against each of the specified principals.
func (doorman *LadonDoorman) ExpandPrincipals(audience string, principals Principals) Principals {
	result := principals[:]

	tags, ok := doorman.tags[audience]
	if !ok {
		return result
	}

	for tag, members := range tags {
		for _, member := range members {
			for _, principal := range principals {
				if principal == member {
					prefixed := fmt.Sprintf("tag:%s", tag)
					result = append(result, prefixed)
				}
			}
		}
	}
	return result
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
