package doorman

import (
	"fmt"

	"github.com/ory/ladon"
	manager "github.com/ory/ladon/manager/memory"
	log "github.com/sirupsen/logrus"
)

const maxInt int64 = 1<<63 - 1

// Tags map tag names to principals.
type Tags map[string]Principals

// LadonDoorman is the backend in charge of checking requests against policies.
type LadonDoorman struct {
	policiesSources []string
	configs         map[string]*Configuration
	_auditLogger    *auditLogger
}

// Configuration represents the policies file content.
type Configuration struct {
	Audience  string
	JWTIssuer string `json:"jwtIssuer"`
	Tags      Tags
	Policies  []*ladon.DefaultPolicy

	ladon        *ladon.Ladon
	jwtValidator JWTValidator
}

// GetTags returns the tags principals for the ones specified.
func (c *Configuration) GetTags(principals Principals) Principals {
	result := Principals{}
	for tag, members := range c.Tags {
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

// New instantiates a new doorman.
func New(policies []string) *LadonDoorman {
	w := &LadonDoorman{
		policiesSources: policies,
		configs:         map[string]*Configuration{},
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
	// First, load each configuration file.
	newConfigs := map[string]*Configuration{}
	for _, source := range doorman.policiesSources {
		configs, err := loadSource(source)
		if err != nil {
			return err
		}
		for _, config := range configs {
			_, exists := newConfigs[config.Audience]
			if exists {
				return fmt.Errorf("duplicated audience %q (source %q)", config.Audience, source)
			}

			if config.JWTIssuer != "" {
				log.Infof("Enable JWT validation from %q", config.JWTIssuer)
				v, err := NewJWTValidator(config.JWTIssuer)
				if err != nil {
					return err
				}
				config.jwtValidator = v
			} else {
				log.Warningf("No JWT verification for %q.", config.Audience)
			}

			config.ladon = &ladon.Ladon{
				Manager:     manager.NewMemoryManager(),
				AuditLogger: doorman.auditLogger(),
			}
			for _, pol := range config.Policies {
				log.Debugf("Load policy %q: %s", pol.GetID(), pol.GetDescription())
				err := config.ladon.Manager.Create(pol)
				if err != nil {
					return err
				}
			}
			newConfigs[config.Audience] = config
		}
	}
	// Only if everything went well, replace existing configs with new ones.
	doorman.configs = newConfigs
	return nil
}

// JWTValidator returns the JWT validator for the specified audience.
func (doorman *LadonDoorman) JWTValidator(audience string) (JWTValidator, error) {
	c, ok := doorman.configs[audience]
	if !ok {
		return nil, fmt.Errorf("unknown audience %q", audience)
	}
	return c.jwtValidator, nil
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

	c, ok := doorman.configs[audience]
	if !ok {
		// Explicitly log denied request using audit logger.
		doorman.auditLogger().logRequest(false, r, ladon.Policies{})
		return false
	}

	// For each principal, use it as the subject and query ladon backend.
	for _, principal := range request.Principals {
		r.Subject = principal
		if err := c.ladon.IsAllowed(r); err == nil {
			return true
		}
	}
	return false
}

// ExpandPrincipals will match the tags defined in the configuration for this audience
// against each of the specified principals.
func (doorman *LadonDoorman) ExpandPrincipals(audience string, principals Principals) Principals {
	c, ok := doorman.configs[audience]
	if !ok {
		return principals
	}

	return append(principals, c.GetTags(principals)...)
}
