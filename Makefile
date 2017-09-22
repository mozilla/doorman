GO_LINT := $(GOPATH)/bin/golint
GO_BINDATA := $(GOPATH)/bin/go-bindata
DATA_FILES := ./utilities/openapi.yaml ./utilities/contribute.yaml

main: utilities/bindata.go *.go utilities/*.go warden/*.go
	CGO_ENABLED=0 go build -o main *.go

clean:
	rm -f main *.coverprofile utilities/bindata.go

$(GO_BINDATA):
	go get github.com/jteeuwen/go-bindata/...

utilities/bindata.go: $(GO_BINDATA) $(DATA_FILES)
	$(GO_BINDATA) -o utilities/bindata.go -pkg utilities $(DATA_FILES)

policies.yaml:
	touch policies.yaml

serve: main policies.yaml
	./main

$(GO_LINT):
	go get github.com/golang/lint/golint

lint: $(GO_LINT)
	$(GO_LINT) . ./utilities ./warden
	go vet . ./utilities ./warden

test: policies.yaml utilities/bindata.go lint
	go test -v -coverprofile=main.coverprofile -coverpkg=. .
	go test -v -coverprofile=warden.coverprofile -coverpkg=./warden ./warden
	go test -v -coverprofile=utilities.coverprofile -coverpkg=./utilities ./utilities
	# Exclude bindata.go from coverage.
	sed -i '/bindata.go/d' utilities.coverprofile

fmt:
	gofmt -w -s *.go ./utilities/*.go ./warden/*.go

docker-build: main
	docker build -t mozilla/iam .

docker-run:
	docker run --name iam --rm mozilla/iam
