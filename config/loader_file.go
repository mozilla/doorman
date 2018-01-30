package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/mozilla/doorman/doorman"
)

// notSpecified is a simple string to detect unspecified values while unmarshalling.
const notSpecified = "N/A"

// FileLoader loads from local disk (file, folder)
type FileLoader struct{}

// CanLoad will return true if the path exists.
func (f *FileLoader) CanLoad(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// Load reads the local file or scans the folder.
func (f *FileLoader) Load(path string) (doorman.ServicesConfig, error) {
	log.Infof("Load %q locally", path)

	// File always exists because CanLoad() returned true.
	fileInfo, _ := os.Stat(path)

	// If path is a folder, list files.
	filenames := []string{path}
	if fileInfo.IsDir() {
		log.Debugf("List files in folder %q", path)
		fileInfos, err := ioutil.ReadDir(path)
		if err != nil {
			return nil, err
		}
		filenames = []string{}
		for _, fileInfo := range fileInfos {
			if fileInfo.IsDir() {
				continue
			}
			filename := filepath.Join(path, fileInfo.Name())
			log.Debugf("Found %q", filename)
			filenames = append(filenames, filename)
		}
	}

	// Load configurations.
	configs := doorman.ServicesConfig{}
	for _, f := range filenames {
		config, err := loadFile(f)
		if err != nil {
			return nil, err
		}
		configs = append(configs, *config)
	}
	return configs, nil
}

func loadFile(filename string) (*doorman.ServiceConfig, error) {
	log.Debugf("Parse file %q", filename)
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if len(fileContent) == 0 {
		return nil, fmt.Errorf("empty file %q", filename)
	}

	config := doorman.ServiceConfig{
		IdentityProvider: notSpecified,
	}
	if err := yaml.Unmarshal(fileContent, &config); err != nil {
		return nil, err
	}
	if config.IdentityProvider == notSpecified {
		return nil, fmt.Errorf("identityProvider not specified in %q", filename)
	}
	config.Source = filename

	return &config, nil
}
