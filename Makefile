.PHONY: setup tidy fmt vet lint test frontend-install frontend-lint frontend-test check build run-api run-ingest

setup: tidy frontend-install

tidy:
	go mod tidy

fmt:
	gofmt -w $(shell find . -name '*.go' -not -path './frontend/*')

vet:
	go vet ./...

lint:
	go test ./... >/dev/null

frontend-install:
	cd frontend && npm install

frontend-lint:
	cd frontend && npm run lint

frontend-test:
	cd frontend && npm run test

test:
	go test ./...

check: fmt vet lint test frontend-lint frontend-test

build:
	go build ./cmd/api ./cmd/ingest
