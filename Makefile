# Detect the operating system
UNAME_S := $(shell uname -s)

build:
	cd rds-kpi-collector-poc && go build -o rds-kpi-collector

# Mac installation via Homebrew
install-golangci-lint-mac:
	brew install golangci-lint

# Linux/CI installation via go install
install-golangci-lint-linux:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.0.1

# Platform-agnostic golangci-lint installation
install-golangci-lint:
ifeq ($(UNAME_S),Darwin)
	$(MAKE) install-golangci-lint-mac
else
	$(MAKE) install-golangci-lint-linux
endif

# Lint depends only on golangci-lint installation
lint: install-golangci-lint
	cd rds-kpi-collector-poc && golangci-lint run --timeout 10m0s