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
}

// Configuration represents the policies file content.
type Configuration struct {
	Audience string
	Tags     Tags
	Policies []*ladon.DefaultPolicy
}

// New instantiates a new doorman.
func New(policies []string, issuer string) (*LadonDoorman, error) {
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
	if err := w.loadPolicies(); err != nil {
		return nil, err
	}
	return w, nil
}

// JWTIssuer returns the URL of the JWT issuer (if configured)
func (doorman *LadonDoorman) JWTIssuer() string {
	return doorman.jwtIssuer
}

// IsAllowed is responsible for deciding if subject can perform action on a resource with a context.
func (doorman *LadonDoorman) IsAllowed(audience string, request *Request) (bool, Principals) {
	l, ok := doorman.ladons[audience]
	if !ok {
		return false, request.Principals
	}

	// Expand principals with local tags.
	tagPrincipals := doorman.lookupTags(audience, request.Principals)
	principals := append(request.Principals, tagPrincipals...)
	// Expand principals with roles.
	if roles, ok := request.Context["roles"]; ok {
		for _, role := range roles.([]string) {
			prefixed := fmt.Sprintf("role:%s", role)
			principals = append(principals, prefixed)
		}
	}

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

	// For each principal, use it as the subject and query ladon backend.
	for _, principal := range principals {
		r.Subject = principal
		if err := l.IsAllowed(r); err == nil {
			return true, principals
		}
	}
	return false, principals
}

// lookupTags will match the tags defined in the configuration for this audience
// against each of the specified principals.
func (doorman *LadonDoorman) lookupTags(audience string, principals Principals) Principals {
	var tags Principals
	for tag, members := range doorman.tags[audience] {
		for _, member := range members {
			for _, principal := range principals {
				if principal == member {
					prefixed := fmt.Sprintf("tag:%s", tag)
					tags = append(tags, prefixed)
				}
			}
		}
	}
	return tags
}

// LoadPolicies (re)loads configuration and policies from the YAML files.
func (doorman *LadonDoorman) loadPolicies() error {
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
		doorman.tags[config.Audience] = config.Tags
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
