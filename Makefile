GIT := git
PRE_COMMIT := pre-commit

JSONNET_VERSION := v0.22.0
GOLANGCI_VERSION := v2.11.4

VERSION := v0.3.2

.PHONY: fmt
fmt:
	@go fmt ./...

.PHONY: lint
lint:
	@golangci-lint run -v ./...

.PHONY: test
test:
	@go test -v ./...

.PHONY: run
run:
	$(PRE_COMMIT) autoupdate
	$(PRE_COMMIT) run --all-files

.PHONY: release
release:
	@$(GIT) fetch --tags --prune
	@$(GIT) tag --sign -m "GPG signed $(VERSION) tag" $(VERSION)
	@$(GIT) push --tags

.PHONY: setup
setup:
	go install github.com/google/go-jsonnet/cmd/jsonnet-lint@$(JSONNET_VERSION)
	go install github.com/google/go-jsonnet/cmd/jsonnetfmt@$(JSONNET_VERSION)

	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)
