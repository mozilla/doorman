main:
	go build main.go *.go

serve:
	go run *.go

docker-build: main
	docker build -t mozilla/iam .

test:
	go test -v ./utilities ./warden

fmt:
	gofmt -w -s *.go ./utilities/*.go ./warden/*.go
