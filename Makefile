MODULE := github.com/gabeduke/civoctl
VERSION = $(shell git rev-parse --short HEAD)

.PHONY: showver doc run build snapshot release fmt

showver:
	$(info $(VERSION))

doc:
	$(info http://localhost:6060/pkg/$(MODULE))
	godoc

run:
	go run main.go run

build:
	go build -o bin/civoctl -ldflags="-s -w -X main.Version=$(VERSION)"

snapshot:
	goreleaser --snapshot --skip-publish --rm-dist

release:
	goreleaser

fmt:
	go fmt ./...