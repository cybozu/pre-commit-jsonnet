PRE_COMMIT := pre-commit

JSONNET_VERSION := v0.18.0
GOLANGCI_VERSION := v1.46.1

.PHONY: fmt
fmt:
	@go fmt ./...

.PHONY: lint
lint:
	@golangci-lint run ./...

.PHONY: test
test:
	@go test -v ./...

.PHONY: run
run:
	$(PRE_COMMIT) autoupdate
	$(PRE_COMMIT) run --all-files

.PHONY: setup
setup:
	@go install github.com/google/go-jsonnet/cmd/jsonnet-lint@$(JSONNET_VERSION)
	@go install github.com/google/go-jsonnet/cmd/jsonnetfmt@$(JSONNET_VERSION)

	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin $(GOLANGCI_VERSION)
