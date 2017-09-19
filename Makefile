main:
	go build main.go *.go

serve:
	go run *.go

docker-build: main
	docker build -t mozilla/iam .

test:
	go vet . ./utilities ./warden
	go test -covermode set -v ./utilities ./warden

fmt:
	gofmt -w -s *.go ./utilities/*.go ./warden/*.go
