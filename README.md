# KPI Collection Tool

Tool to automate metrics gathering and visualization for RDS KPIs in disconnected environments.

## Building

### Using Make (recommended)

```bash
# Build the binary
make build

# This creates a binary named 'kpi-collector' in the project root
```

### Using Go directly

```bash
go build -o kpi-collector ./cmd/rds-kpi-collector
```

## Running

The tool can run in two modes:

### 1. Using Kubeconfig (Automatic Discovery)

Automatically discovers Thanos URL and creates a service account token:

```bash
./kpi-collector --cluster-name my-cluster --kubeconfig ~/.kube/config
```

### 2. Using Manual Credentials

Provide Thanos URL and bearer token directly:

```bash
./kpi-collector \
  --cluster-name my-cluster \
  --token YOUR_BEARER_TOKEN \
  --thanos-url thanos-querier.example.com
```

## Command Line Flags

| Flag | Required | Description |
|------|----------|-------------|
| `--cluster-name` | Yes | Name of the cluster being monitored |
| `--kubeconfig` | No* | Path to kubeconfig file for auto-discovery |
| `--token` | No* | Bearer token for Thanos authentication |
| `--thanos-url` | No* | Thanos querier URL (without https://) |
| `--insecure-tls` | No | Skip TLS certificate verification (dev only) |

\* Either provide `--kubeconfig` OR both `--token` and `--thanos-url`
