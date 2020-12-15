VERSION=v0.1.0

build:
	go build -o ./bin/dns-update  -ldflags="-s -w -X main.version=$(VERSION)" -trimpath ./cmd/dns-update

test:
	go test ./cmd/dns-update
