package doorman

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"

	log "github.com/sirupsen/logrus"
)

type headers map[string]string

type githubLoader struct {
	headers headers
}

func (ghl *githubLoader) CanLoad(url string) bool {
	regexpRepo, _ := regexp.Compile("^https://.*github.*/.*$")
	return regexpRepo.MatchString(url)
}

func (ghl *githubLoader) Load(url string) ([]*Configuration, error) {
	log.Infof("Load %q from Github", url)

	regexpFile, _ := regexp.Compile("^.*\\.ya?ml$")

	urls := []string{}
	// Single file URL.
	if regexpFile.MatchString(url) {
		urls = []string{url}
	} else {
		// Folder on remote repo.
		return nil, fmt.Errorf("loading from Github folder is not supported yet")
	}
	// Load configurations.
	configs := []*Configuration{}
	for _, url := range urls {
		tmpFile, err := download(url, ghl.headers)
		if err != nil {
			return nil, err
		}
		config, err := loadFile(tmpFile.Name())
		if err != nil {
			return nil, err
		}

		// Only delete temp file if successful
		os.Remove(tmpFile.Name())
		configs = append(configs, config)
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

func init() {
	githubToken := os.Getenv("GITHUB_TOKEN")

	loaders = append(loaders, &githubLoader{
		headers: headers{
			"Authorization": fmt.Sprintf("token %s", githubToken),
		},
	})
}
