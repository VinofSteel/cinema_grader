.DEFAULT_GOAL := build

fmt:
	gofmt -w .
.PHONY:fmt

test:
	go test ./... -v -count=1
.PHONY: test

integration-test:
	go test ./tests/ -v -count=1
.PHONY: integration-test

build: fmt
	air
.PHONY:build