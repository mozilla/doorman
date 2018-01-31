GO_LINT := $(GOPATH)/bin/golint
GO_DEP := $(GOPATH)/bin/dep
GO_BINDATA := $(GOPATH)/bin/go-bindata
GO_PACKAGE := $(GOPATH)/src/github.com/mozilla/doorman
DATA_FILES := ./api/openapi.yaml ./api/contribute.yaml
SRC := *.go ./config/*.go ./api/*.go ./authn/*.go ./doorman/*.go
PACKAGES := ./ ./config/ ./api/ ./authn/ ./doorman/

main: vendor api/bindata.go $(SRC) $(GO_PACKAGE)
	CGO_ENABLED=0 go build -o main *.go

clean:
	rm -f main coverage.txt api/bindata.go vendor

$(GOPATH):
	mkdir -p $(GOPATH)

$(GO_PACKAGE): $(GOPATH)
	mkdir -p $(shell dirname ${GO_PACKAGE})
	if [ ! -e $(GOPACKAGE) ]; then ln -sf $$PWD $(GO_PACKAGE); fi

$(GO_DEP): $(GOPATH) $(GO_PACKAGE)
	go get -u github.com/golang/dep/cmd/dep

$(GO_BINDATA): $(GOPATH)
	go get github.com/jteeuwen/go-bindata/...

vendor: $(GO_DEP) Gopkg.lock Gopkg.toml
	$(GO_DEP) ensure

api/bindata.go: $(GO_BINDATA) $(DATA_FILES)
	$(GO_BINDATA) -o api/bindata.go -pkg api $(DATA_FILES)

policies.yaml:
	touch policies.yaml

serve: main policies.yaml
	./main

$(GO_LINT):
	go get github.com/golang/lint/golint

lint: $(GO_LINT)
	$(GO_LINT) $(PACKAGES)
	go vet $(PACKAGES)

fmt:
	gofmt -w -s $(SRC)

test: vendor policies.yaml api/bindata.go lint
	go test -v $(PACKAGES)

test-coverage: vendor policies.yaml api/bindata.go
	# Multiple package coverage script from https://github.com/pierrre/gotestcover
	echo 'mode: atomic' > coverage.txt && go list ./... | grep -v /vendor/ | xargs -n1 -I{} sh -c 'go test -v -covermode=atomic -coverprofile=coverage.tmp {} && tail -n +2 coverage.tmp >> coverage.txt' && rm coverage.tmp
	# Exclude bindata.go from coverage.
	sed -i '/bindata.go/d' coverage.txt

docker-build: main
	docker build -t mozilla/doorman .

docker-run:
	docker run --name doorman --rm mozilla/doorman

.venv/bin/sphinx-build:
	virtualenv .venv
	.venv/bin/pip install -r docs/requirements.txt

docs: .venv/bin/sphinx-build docs/*.rst
	.venv/bin/sphinx-build -a -W -n -b html -d docs/_build/doctrees docs docs/_build/html
	@echo
	@echo "Build finished. The HTML pages are in $(SPHINX_BUILDDIR)/html/index.html"

docs-publish: docs
	# https://github.com/tschaub/gh-pages
	gh-pages -d api-docs
