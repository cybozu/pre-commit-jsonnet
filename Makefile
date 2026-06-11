GIT := git
PRE_COMMIT := pre-commit

VERSION := v0.3.3

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
	go install tool
