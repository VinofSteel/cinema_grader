.DEFAULT_GOAL := build

fmt:
	gofmt -w .
.PHONY:fmt

test:
	go test -v ./...
.PHONY: test

integration-test:
	go test ./tests/ -v -count=1

build: fmt
	air
.PHONY:build