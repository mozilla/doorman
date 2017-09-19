main:
	go build main.go *.go

serve:
	go run *.go

docker-build: main
	docker build -t mozilla/iam .

test:
	go vet . ./utilities ./warden
	go test -v -coverprofile=utilities.coverprofile -coverpkg=./utilities ./utilities
	go test -v -coverprofile=warden.coverprofile -coverpkg=./warden ./warden

fmt:
	gofmt -w -s *.go ./utilities/*.go ./warden/*.go
