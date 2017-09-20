main:
	go build -o main *.go

serve:
	go run *.go

docker-build: main
	docker build -t mozilla/iam .

docker-run:
	docker run --name iam --rm mozilla/iam

test:
	go vet . ./utilities ./warden
	go test -v -coverprofile=utilities.coverprofile -coverpkg=./utilities ./utilities
	go test -v -coverprofile=warden.coverprofile -coverpkg=./warden ./warden

fmt:
	gofmt -w -s *.go ./utilities/*.go ./warden/*.go
