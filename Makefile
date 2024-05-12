.DEFAULT_GOAL := build

fmt:
	gofmt -w .
.PHONY:fmt

test: fmt
	go test ./... -count=1
.PHONY: test

integration-test: fmt
	go test ./tests/ -count=1
.PHONY: integration-test

build: fmt
	air
.PHONY:build