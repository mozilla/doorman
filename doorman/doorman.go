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

// Doorman is the backend in charge of checking requests against policies.
type Doorman struct {
	PoliciesFilenames []string
	JWTIssuer         string
	ladons            map[string]ladon.Ladon
	groups            map[string][]UserGroup
}

// Principals represent a user (userid, email, groups, ...)
type Principals []string

// Configuration represents the policies file content.
type Configuration struct {
	Audience string
	Groups   map[string]Principals
	Policies []*ladon.DefaultPolicy
}

// UserGroup is a group of principals.
type UserGroup struct {
	Name    string
	Members Principals
}

// Request is the authorization request.
type Request struct {
	// Principals are strings that identify the user.
	Principals Principals
	// Resource is the resource that access is requested to.
	Resource string
	// Action is the action that is requested on the resource.
	Action string
	// Context is the request's environmental context.
	Context ladon.Context
}

// New instantiates a new doorman.
func New(filenames []string, issuer string) (*Doorman, error) {
	// If not specified, read default file in current directory `./policies.yaml`
	if len(filenames) == 0 {
		here, _ := os.Getwd()
		filename := filepath.Join(here, DefaultPoliciesFilename)
		filenames = []string{filename}
	}

	w := &Doorman{
		PoliciesFilenames: filenames,
		JWTIssuer:         issuer,
		ladons:            map[string]ladon.Ladon{},
		groups:            map[string][]UserGroup{},
	}
	if err := w.loadPolicies(); err != nil {
		return nil, err
	}
	return w, nil
}

// IsAllowed is responsible for deciding if subject can perform action on a resource with a context.
func (doorman *Doorman) IsAllowed(audience string, request *Request) (bool, Principals) {
	l, ok := doorman.ladons[audience]
	if !ok {
		return false, request.Principals
	}

	// Expand principals with local groups.
	groupPrincipals := doorman.lookupGroups(audience, request.Principals)
	principals := append(request.Principals, groupPrincipals...)
	// XXX: expand with request roles.
	// XXX: expand with specific checks.

	// For each principal, use it as the subject and query ladon backend.
	for _, principal := range principals {
		r := &ladon.Request{
			Subject:  principal,
			Resource: request.Resource,
			Action:   request.Action,
			Context:  request.Context,
		}
		if err := l.IsAllowed(r); err == nil {
			return true, principals
		}
	}
	return false, principals
}

// lookupGroups will match the groups defined in the configuration for this audience
// against each of the specified principals.
func (doorman *Doorman) lookupGroups(audience string, principals Principals) Principals {
	var groups Principals
	for _, group := range doorman.groups[audience] {
		for _, member := range group.Members {
			for _, principal := range principals {
				if principal == member {
					prefixed := fmt.Sprintf("group:%s", group.Name)
					groups = append(groups, prefixed)
				}
			}
		}
	}
	return groups
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

		var groups []UserGroup
		for name, members := range config.Groups {
			log.Infof("Load group %q", name)
			groups = append(groups, UserGroup{name, members})
		}
		doorman.groups[config.Audience] = groups
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
