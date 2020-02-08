MODULE := github.com/gabeduke/civoctl

doc:
	$(info http://localhost:6060/pkg/$(MODULE))
	godoc

run:
	go run main.go run

build:
	go build -o bin/civoctl

fmt:
	go fmt ./...