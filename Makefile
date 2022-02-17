# Makefile for u9k

.PHONY: server
server: # server creates the u9k server binary
	CGO_ENABLED=0 go build ./cmd/server

.PHONY: container
container:
	podman image build -t docker.io/jacksgt/u9k:dev .

.PHONY: test
go-tests: # runs all the tests (requires DB / object store)
	. ./auth/dev-env.sh && go test ./...

.PHONY: go-unit-tests
go-unit-tests:
	go test -test.short ./...

.PHONY: js-lint
js-lint:
	( PATH=./node_modules/.bin/:$$PATH node_modules/.bin/eslint static/js/ )

.PHONY: dev # runs the development version
dev:
	. ./auth/dev-env.sh && go run cmd/server/main.go -reloadTemplates=true

.PHONY: debug # runs the development version with delve debugger
debug:
	. ./auth/dev-env.sh && dlv debug cmd/server/main.go
