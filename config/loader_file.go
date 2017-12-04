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

	var config doorman.ServiceConfig
	if err := yaml.Unmarshal(fileContent, &config); err != nil {
		return nil, err
	}
	config.Source = filename

	if config.Service == "" {
		return nil, fmt.Errorf("empty service in %q", filename)
	}

	if len(config.Policies) == 0 {
		log.Warningf("No policies found in %q", filename)
	}

	log.Infof("Found service %q", config.Service)
	log.Infof("Found %d tags", len(config.Tags))
	log.Infof("Found %d policies", len(config.Policies))

	return &config, nil
}
