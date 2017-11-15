package doorman

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/leplatrem/iam/utilities"
)

type fileLoader struct{}

func (f *fileLoader) CanLoad(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (f *fileLoader) Load(path string) ([]*Configuration, error) {
	log.Infof("Load %q locally", path)

	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, err
	}
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
	configs := []*Configuration{}
	for _, f := range filenames {
		config, err := loadFile(f)
		if err != nil {
			return nil, err
		}
		configs = append(configs, config)
	}
	return configs, nil
}

func loadFile(filename string) (*Configuration, error) {
	log.Debugf("Parse file %q", filename)
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
		log.Warningf("No policies found in %q", filename)
	}

	log.Infof("Found audience %q", config.Audience)
	log.Infof("Found %d tags", len(config.Tags))
	log.Infof("Found %d policies", len(config.Policies))

	return &config, nil
}

func init() {
	loaders = append(loaders, &fileLoader{})
}
