package doorman

import (
	"fmt"
)

// Loader is responsible for loading the policies from files, URLs, etc.
type Loader interface {
	CanLoad(source string) bool
	Load(source string) ([]*ServiceConfig, error)
}

var loaders []Loader

func init() {
	loaders = []Loader{}
}

func loadSource(source string) ([]*ServiceConfig, error) {
	for _, loader := range loaders {
		if loader.CanLoad(source) {
			return loader.Load(source)
		}
	}
	return nil, fmt.Errorf("no appropriate loader found for %q", source)
}
