package config

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"

	log "github.com/sirupsen/logrus"

	"github.com/mozilla/doorman/doorman"
)

type headers map[string]string

// GithubLoader reads configuration from Github URLs.
type GithubLoader struct {
	Token string
}

// CanLoad will return true if the URL contains github
func (ghl *GithubLoader) CanLoad(url string) bool {
	regexpRepo, _ := regexp.Compile("^https://.*github.*/.*$")
	return regexpRepo.MatchString(url)
}

// Load downloads the URL into a temporary folder and loads it from disk
func (ghl *GithubLoader) Load(source string) (doorman.ServicesConfig, error) {
	log.Infof("Load %q from Github", source)

	regexpFile, _ := regexp.Compile("^.*\\.ya?ml$")

	urls := []string{}
	// Single file URL.
	if regexpFile.MatchString(source) {
		urls = []string{source}
	} else {
		// Folder on remote repo.
		return nil, fmt.Errorf("loading from Github folder is not supported yet")
	}

	headers := headers{
		"Authorization": fmt.Sprintf("token %s", ghl.Token),
	}

	// Load configurations.
	configs := doorman.ServicesConfig{}
	for _, url := range urls {
		tmpFile, err := download(url, headers)
		if err != nil {
			return nil, err
		}
		config, err := loadFile(tmpFile.Name())
		if err != nil {
			return nil, err
		}
		config.Source = url

		// Only delete temp file if successful
		os.Remove(tmpFile.Name())
		configs = append(configs, *config)
	}
	return configs, nil
}

func download(url string, headers headers) (*os.File, error) {
	f, err := ioutil.TempFile("", "doorman-policy-")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	log.Debugf("Download %q", url)
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	size, err := io.Copy(f, response.Body)
	if err != nil {
		return nil, err
	}
	log.Debugf("Downloaded %dkB", size/1000)
	return f, nil
}
