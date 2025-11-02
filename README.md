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

With custom sampling parameters:

```bash
./kpi-collector \
  --cluster-name my-cluster \
  --kubeconfig ~/.kube/config \
  --frequency 30 \
  --duration 1h \
  --output my-metrics.json \
  --log my-app.log
```

### 2. Using Manual Credentials

Provide Thanos URL and bearer token directly:

```bash
./kpi-collector \
  --cluster-name my-cluster \
  --token YOUR_BEARER_TOKEN \
  --thanos-url thanos-querier.example.com
```

With custom sampling parameters:

```bash
./kpi-collector \
  --cluster-name my-cluster \
  --token YOUR_BEARER_TOKEN \
  --thanos-url thanos-querier.example.com \
  --frequency 120 \
  --duration 30m \
  --output results.json
```

## Command Line Flags

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--cluster-name` | Yes | - | Name of the cluster being monitored |
| `--kubeconfig` | No* | - | Path to kubeconfig file for auto-discovery |
| `--token` | No* | - | Bearer token for Thanos authentication |
| `--thanos-url` | No* | - | Thanos querier URL (without https://) |
| `--insecure-tls` | No | false | Skip TLS certificate verification (dev only) |
| `--frequency` | No | 60 | Sampling frequency in seconds |
| `--duration` | No | 45m | Total duration for sampling (e.g. 10s, 1m, 2h) |
| `--output` | No | kpi-output.json | Output file name for results |
| `--log` | No | kpi.log | Log file name |

\* Either provide `--kubeconfig` OR both `--token` and `--thanos-url`
