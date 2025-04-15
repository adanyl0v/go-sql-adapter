# Tests and benchmarks
# --------------------

.PHONY: test
test: lint
	go test -v ./...

.PHONY: cover
cover: lint
	go test -v -coverprofile cover.out ./...
	go tool cover -html cover.out -o cover.html
	rm cover.out

.PHONY: bench
bench: lint
	go test -v -bench ./...

# Linting and formatting
# ----------------------

.PHONY: tidy
tidy:
	go mod tidy
	go vet ./...

.PHONY: lint-verify
lint-verify: tidy
	golangci-lint config verify

.PHONY: lint
lint: lint-verify
	golangci-lint run ./...

.PHONY: lint-fix
lint-fix: lint-verify
	golangci-lint run ./... --fix

# Mocks
# -----

.PHONY: mocks
mocks:
	go generate ./...
