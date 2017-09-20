GO_BINDATA := $(GOPATH)/bin/go-bindata
DATA_FILES := ./utilities/openapi.yaml ./utilities/contribute.yaml

main: utilities/bindata.go *.go utilities/*.go warden/*.go
	CGO_ENABLED=0 go build -o main *.go

$(GO_BINDATA):
	go get github.com/jteeuwen/go-bindata/...

utilities/bindata.go: $(GO_BINDATA) $(DATA_FILES)
	$(GO_BINDATA) -o utilities/bindata.go -pkg utilities $(DATA_FILES)

serve:
	go run *.go

docker-build: main
	docker build -t mozilla/iam .

docker-run:
	docker run --name iam --rm mozilla/iam

test: utilities/bindata.go
	go vet . ./utilities ./warden
	go test -v -coverprofile=utilities.coverprofile -coverpkg=./utilities ./utilities
	go test -v -coverprofile=warden.coverprofile -coverpkg=./warden ./warden

fmt:
	gofmt -w -s *.go ./utilities/*.go ./warden/*.go
