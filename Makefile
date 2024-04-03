.DEFAULT_GOAL := build

fmt:
	gofmt -w .
.PHONY:fmt

build: fmt
	air
.PHONY:build