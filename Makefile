MODULE_PKG := github.com/NethermindEth/aztec-p2p-explorer
BIN_NAME ?= aztec-p2p-explorer

# If the environment has GITHUB_SHA we're running in a Github action, otherwise we try to
# get the commit hash from git. This avoids having to copy .git into docker build context.
GIT_HASH := $(or $(shell echo $(GITHUB_SHA)),$(shell git rev-parse --short HEAD))
BUILD_DATE := $(shell date -u '+%Y%m%dT%H%M%S')
LDFLAGS := -ldflags="-s -w -X $(MODULE_PKG)/build.BuildDate=${BUILD_DATE} -X ${MODULE_PKG}/build.Commit=${GIT_HASH}"

BIN_DIR ?= $(shell pwd)/bin

.PHONY: default
default: build

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

ci-full: clean lint test security
.PHONY: ci-full

ci: clean lint test
.PHONY: ci

clean:
	@rm -f aztec-p2p-explorer
	@rm -fr ./dist
.PHONY: clean

.PHONY: security
security: ## Run security checks
	@ echo "▶️ gosec ./..."
	@ gosec  ./...
	@ echo "✅ gosec golint ./..."

.PHONY: lint
LINT_ARGS ?= -v --build-tags production
LINT_TARGETS ?= ./...
lint: ## Lint Go code with the installed golangci-lint
	@ echo "▶️ golangci-lint run $(LINT_ARGS) $(LINT_TARGETS)"
	golangci-lint run $(LINT_ARGS) $(LINT_TARGETS)
	@ echo "✅ golangci-lint run"

.PHONY: vet
VET_ARGS ?= -tags production
VET_TARGETS ?= ./...
vet:
	go vet $(VET_ARGS) $(VET_TARGETS)

.PHONY: staticcheck
STATICCHECK_TARGETS ?= ./...
STATICCHECK_ARGS ?= -tags production
staticcheck: ## Run staticcheck linter
	@ echo "▶️ gstaticcheck $(STATICCHECK_ARGS) $(STATICCHECK_TARGETS)"
	CGO_ENABLED=0 staticcheck $(STATICCHECK_ARGS) $(STATICCHECK_TARGETS)
	@ echo "✅ staticcheck $(STATICCHECK_ARGS) $(STATICCHECK_TARGETS)"

.PHONY: test
TEST_TARGETS ?= ./...
TEST_ARGS ?= -v -coverprofile=coverage.txt -tags production
test: ## Test the Go modules within this package.
	@ echo ▶️ go test $(TEST_ARGS) $(TEST_TARGETS)
	BUNDEBUG=1 go test $(TEST_ARGS) $(TEST_TARGETS)
	@ echo ✅ success!

	@ echo ▶️ go tool cover -func=coverage.txt
	go tool cover -func=coverage.txt
	@ echo ✅ success!

	@ echo ▶️ go tool cover -html=coverage.txt -o cover.html
	go tool cover -html=coverage.txt -o cover.html
	@ echo ✅ success!

.PHONY: test-debug
test-debug: ## Test the Go modules within this package.
	@ echo ▶️ go test $(TEST_ARGS) $(TEST_TARGETS)
	BUNDEBUG=2 go test $(TEST_ARGS) $(TEST_TARGETS)
	@ echo ✅ success!

	@ echo ▶️ go tool  cover -func=coverage.txt
	go tool cover -func=coverage.txt
	@ echo ✅ success!

	@ echo ▶️ go tool cover -html=coverage.txt -o cover.html
	go tool cover -html=coverage.txt -o cover.html
	@ echo ✅ success!

.PHONY: bench
bench: ## Run benchmarks
	@ echo ▶️ go test -bench=. -benchmem
	go test -bench=. -benchmem ./...
	@ echo ✅ success!

build: build-development
.PHONY: build

build-development: docs ## Build the Go binary
	go generate -v ./...
	go build $(LDFLAGS) -tags development -o $(BIN_NAME)
.PHONY: build-development

build-production:
	go build $(LDFLAGS) -tags production -o $(BIN_NAME)
.PHONY: build-production

docker:
	@ docker buildx build --build-arg GITHUB_SHA=${GITHUB_SHA} --target aztec-p2p-explorer-runtime -t nethermindeth/aztec-p2p-explorer .

.PHONY: certs
certs:
	@echo ▶️ generating self signed certs
	@mkdir -p certs
	@openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout certs/key.pem -out certs/cert.pem -subj "/C=US/ST=CA/L=San Francisco/O=NethermindEth/OU=IT Department/CN=localhost"
	@ echo ✅ success!


INSTALL_CMD := $$(if [ "$$(uname)" = "Linux" ]; then if [ -x "$$(command -v apt-get)" ]; then echo "sudo apt-get -y"; elif [ -x "$$(command -v dnf)" ]; then echo "sudo dnf -y"; elif [ -x "$$(command -v pacman)" ]; then echo "sudo pacman -y"; elif [ -x "$$(command -v yum)" ]; then echo "sudo yum -y"; elif [ -x "$$(command -v zypper)" ]; then echo "sudo zypper -y"; else echo "Unsupported package manager. Please install manually:"; fi; elif [ "$$(uname)" = "Darwin" ]; then echo "brew install"; else echo "Unsupported OS.  Please install manually:"; fi)
.PHONY: base-deps
base-deps:
	@ echo ▶️ checking for go installation
	@if ! command -v go >/dev/null 2>&1; then \
        echo "Go is not installed."; \
		echo "You could install it by running the following commands:"; \
		echo "$(INSTALL_CMD) go"; \
		echo "export PATH=\$$PATH:$(shell go env GOPATH)/bin"; \
		echo "export PATH=\$$PATH:$(GOPATH)/bin | sudo tee -a /etc/profile"; \
		exit 1; \
	fi
	@ echo ✅ success!

.PHONY: dev-deps
dev-deps: ## Install develpment tools
	@ echo starting
	@ echo ✅ success!
	@ echo ▶️ install swaggo/swag
	@if ! command -v swag >/dev/null 2>&1; then \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	@ echo ✅ success!
	@ echo ▶️ install golangci-lint
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@ echo ✅ success!
	@ echo ▶️ install staticcheck
	@if ! command -v staticcheck >/dev/null 2>&1; then \
		go install honnef.co/go/tools/cmd/staticcheck@latest; \
	fi
	@ echo ✅ success!
	@ echo ▶️ install gosec
	@if ! command -v gosec >/dev/null 2>&1; then \
		go install github.com/securego/gosec/cmd/gosec@latest; \
	fi
	@ echo ✅ success!
	@ echo ▶️ install go-swagger
	@ go install github.com/go-swagger/go-swagger/cmd/swagger@latest
	@ echo ✅ success!
	@ echo ▶️ install sqlboiler
	@ go install github.com/aarondl/sqlboiler/v4@latest
	@ echo ✅ success!
	@ echo ▶️ install sqlboiler-psql
	@ go install github.com/aarondl/sqlboiler/v4/drivers/sqlboiler-psql@latest
	@ echo ✅ success!

.PHONY: deps
deps: dev-deps base-deps

.PHONY: docs
docs: ## Generate API documentation
	@ echo ▶️ generating swagger docs
	@ swag init -g server/server.go
	@ echo ▶️ generating markdown docs
	@ swagger generate markdown -f docs/swagger.json --output=./docs/api.md
	@ echo ✅ success!

.PHONY: models
models:
	sqlboiler --no-tests --no-hooks psql

.PHONY: migrate-up
migrate-up:
	migrate -database 'postgres://explorer:explorer@localhost:15432/explorer?sslmode=disable' -path database/migrations up

.PHONY: migrate-down
migrate-down:
	migrate -database 'postgres://explorer:explorer@localhost:15432/explorer?sslmode=disable' -path database/migrations down
