.DEFAULT_GOAL := run

fmt:
	gofmt -w .
.PHONY:fmt

test: fmt
	go test ./... -count=1
.PHONY: test

integration-test: fmt
	go test ./tests/ -count=1
.PHONY: integration-test

build: test
	go build -o ./cmd/c_grader/c_grader.exe ./cmd/c_grader/main.go
.PHONY:build

run: fmt
	air
.PHONY:run