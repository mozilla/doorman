package config

import (
	"fmt"
	"strings"
	log "github.com/sirupsen/logrus"

	"github.com/mozilla/doorman/doorman"
)

// lintConfigs inspects the service configuration and warns or returns an error
// if something looks wrong.
func lintConfigs(configs ...doorman.ServiceConfig) error {
	for _, config := range configs {

		if config.Service == "" {
			return fmt.Errorf("empty service in %q", config.Source)
		}

		if len(config.Policies) == 0 {
			log.Warningf("No policies found in %q", config.Source)
		} else {
			log.Infof("Found %d policies", len(config.Policies))
		}

		log.Infof("Found service %q", config.Service)
		log.Infof("Found %d tags", len(config.Tags))

		for _, policy := range config.Policies {
			// HTTP verbs as actions in policies.
			for _, action := range policy.Actions {
				if strings.Contains("get,put,post,delete", strings.ToLower(action)) {
					log.Warningf("Avoid coupling of actions with HTTP verbs (%q in %q)", policy.ID, config.Source)
				}
			}
			// URLs in resources
			for _, resource := range policy.Resources {
				if strings.HasPrefix(resource, "/") {
					log.Warningf("Avoid coupling of resources with API URIs (%q in %q)", policy.ID, config.Source)
				}
			}
		}
	}
	return nil
}
