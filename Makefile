# =========================
# Global Variables
# =========================
GRAFANA_VERSION ?= latest
GRAFANA_PORT ?= 3000

# Detect the operating system
UNAME_S := $(shell uname -s)
# Binary name
BINARY_NAME=kpi-collector

build:
	go build -o $(BINARY_NAME) ./cmd/kpi-collector

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
	golangci-lint run --timeout 10m0s

# Run all tests
test:
	go test ./... -v


install-grafana:
	docker rm -f grafana-kpi || true
	docker run -d \
		--name grafana-kpi \
		-p $(GRAFANA_PORT):3000 \
		-v $(PWD)/grafana/datasource:/etc/grafana/provisioning/datasources:ro \
		-v $(PWD)/grafana/provisioning/dashboards:/etc/grafana/provisioning/dashboards:ro \
		-v $(PWD)/grafana/dashboard:/var/lib/grafana/dashboards:ro \
		-v $(PWD)/collected-data/kpi_metrics.db:/var/lib/grafana/kpi_metrics.db:ro \
		-e "GF_INSTALL_PLUGINS=frser-sqlite-datasource" \
		grafana/grafana:$(GRAFANA_VERSION)


# Install kpi-collector to user's Go bin directory
install:
	go install ./cmd/kpi-collector
	echo "✓ Installed to $(HOME)/go/bin/$(BINARY_NAME)"

# Uninstall kpi-collector
uninstall:
	rm -f $(HOME)/go/bin/$(BINARY_NAME)
	echo "✓ Uninstalled"
