// Package config is in charge of loading policies files from disk or remote Github URL.
//
// It also contains the view for the __reload__ endpoint.
package config

import (
	"fmt"

	"github.com/mozilla/doorman/doorman"
)

// Loader is responsible for loading the policies from files, URLs, etc.
type Loader interface {
	// CanLoad determines if the loader can handle this source.
	CanLoad(source string) bool
	// Parse and return the configs for this source.
	Load(source string) (doorman.ServicesConfig, error)
}

var loaders []Loader

func init() {
	loaders = []Loader{}
}

// AddLoader allows to plug new kinds of loaders.
func AddLoader(l Loader) {
	loaders = append(loaders, l)
}

// Load will load and parse the specified sources.
func Load(sources []string) (doorman.ServicesConfig, error) {
	configs := doorman.ServicesConfig{}
	for _, source := range sources {
		loaded := false
		for _, loader := range loaders {
			if loader.CanLoad(source) {
				c, err := loader.Load(source)
				if err != nil {
					return nil, err
				}
				configs = append(configs, c...)
				loaded = true
			}
		}
		if !loaded {
			return nil, fmt.Errorf("no appropriate loader found for %q", source)
		}
	}
	return configs, nil
}
