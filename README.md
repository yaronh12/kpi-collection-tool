# KPI Collection Tool

Tool to automate metrics gathering and visualization for KPIs in disconnected environments.

## Installation

### Using Make (recommended)

```bash
# Build and install globally
make install

# This installs kpi-collector to ~/go/bin/
# Make sure ~/go/bin is in your PATH
```

To add `~/go/bin` to your PATH, add this to your `~/.zshrc` (macOS) or `~/.bashrc` (Linux):
```bash
export PATH="$HOME/go/bin:$PATH"
```

Then reload your shell:
```bash
source ~/.zshrc  # or source ~/.bashrc on Linux
```

### Uninstall

```bash
make uninstall
```

## Usage

The tool uses subcommands for different operations. Get help anytime with:

```bash
kpi-collector --help
kpi-collector run --help
kpi-collector version
```

## Collecting KPI Metrics

The `run` command gathers KPI metrics from Prometheus/Thanos and stores them in a database.

### Authentication Modes

#### 1. Using Kubeconfig (Automatic Discovery)

Automatically discovers Thanos URL and creates a service account token.

**Basic usage (uses SQLite by default):**
```bash
kpi-collector run \
  --cluster-name my-cluster \
  --kubeconfig ~/.kube/config
```

**With custom sampling parameters:**
```bash
kpi-collector run \
  --cluster-name my-cluster \
  --cluster-type ran \
  --kubeconfig ~/.kube/config \
  --frequency 30 \
  --duration 1h \
  --output my-metrics.json \
  --log my-app.log
```

**Explicitly using SQLite:**
```bash
kpi-collector run \
  --cluster-name my-cluster \
  --kubeconfig ~/.kube/config \
  --db-type sqlite
```

**Using PostgreSQL:**
```bash
kpi-collector run \
  --cluster-name my-cluster \
  --kubeconfig ~/.kube/config \
  --db-type postgres \
  --postgres-url "postgresql://myuser:mypass@localhost:5432/kpi_metrics?sslmode=disable"
```

#### 2. Using Manual Credentials

Provide Thanos URL and bearer token directly.

**Basic usage (uses SQLite by default):**
```bash
kpi-collector run \
  --cluster-name my-cluster \
  --token YOUR_BEARER_TOKEN \
  --thanos-url thanos-querier.example.com
```

**With custom sampling parameters:**
```bash
kpi-collector run \
  --cluster-name my-cluster \
  --token YOUR_BEARER_TOKEN \
  --thanos-url thanos-querier.example.com \
  --frequency 120 \
  --duration 30m \
  --output results.json
```

**Using PostgreSQL:**
```bash
kpi-collector run \
  --cluster-name my-cluster \
  --token YOUR_BEARER_TOKEN \
  --thanos-url thanos-querier.example.com \
  --db-type postgres \
  --postgres-url "postgresql://myuser:mypass@localhost:5432/kpi_metrics?sslmode=disable"
```
## --insecure-tls

Use this flag when running the tool against clusters or Prometheus/Thanos servers with self-signed or untrusted certificates.

```bash
./kpi-collector \
  --cluster-name my-cluster \
  --kubeconfig ~/.kube/config \
  --insecure-tls
```  
### What it does
- Skips TLS certificate verification for all HTTPS requests:
- Kubernetes API calls
- Thanos/Prometheus queries
- Allows execution in environments where the certificate cannot be validated.

### When to use
- Self-signed certificates
- Disconnected / lab / air-gapped clusters
- kubeconfig without a valid CA
- TLS errors such as:
  x509: certificate signed by unknown authority

### Complete Examples

**Development setup with SQLite (default):**
```bash
kpi-collector run \
  --cluster-name dev-cluster \
  --kubeconfig ~/.kube/config \
  --frequency 60 \
  --duration 1h \
  --insecure-tls
```

**Production setup with PostgreSQL:**
```bash
kpi-collector run \
  --cluster-name prod-cluster \
  --token YOUR_BEARER_TOKEN \
  --thanos-url thanos-querier.prod.example.com \
  --db-type postgres \
  --postgres-url "postgresql://kpi_user:secure_password@postgres.example.com:5432/kpi_metrics?sslmode=require" \
  --frequency 30 \
  --duration 24h \
  --output prod-metrics.json \
  --log prod-kpi.log
```

**Using a custom KPIs configuration file:**
```bash
kpi-collector run \
  --cluster-name my-cluster \
  --kubeconfig ~/.kube/config \
  --kpis-file /path/to/custom-kpis.json
```

## Command Line Flags

### Collect Command Flags

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--cluster-name` | Yes | - | Name of the cluster being monitored |
| `--cluster-type` | No | - | Cluster type for categorization: `ran`, `core`, or `hub` |
| `--kubeconfig` | No* | - | Path to kubeconfig file for auto-discovery |
| `--token` | No* | - | Bearer token for Thanos authentication |
| `--thanos-url` | No* | - | Thanos querier URL (without https://) |
| `--insecure-tls` | No | false | Skip TLS certificate verification (dev only) |
| `--frequency` | No | 60 | Sampling frequency in seconds (how often to collect metrics) |
| `--duration` | No | 45m | Total duration for sampling (e.g. 10s, 1m, 2h, 24h) |
| `--output` | No | kpi-output.json | Output file name for results |
| `--log` | No | kpi.log | Log file name |
| `--db-type` | No | sqlite | Database type: `sqlite` or `postgres` |
| `--postgres-url` | No** | - | PostgreSQL connection string |
| `--kpis-file` | No | configs/kpis.json | Path to KPIs configuration file |
| `--grafana-file` | No | - | Path to exported Grafana dashboard JSON to analyze |
| `--summarize` | No | false | Run Grafana AI summarization after KPI collection |
| `--ollama-model` | No | llama3.2:latest | Local Ollama model to use for AI analysis |

\* Either provide `--kubeconfig` OR both `--token` and `--thanos-url`

\*\* Required when `--db-type=postgres`

### Understanding Frequency and Duration

The `--frequency` and `--duration` flags work together to control how metrics are collected:

- **`--frequency`**: How often (in seconds) to collect metrics
  - Example: `--frequency 60` means collect metrics every 60 seconds (once per minute)
  - Lower values = more frequent sampling = more data points
  - Higher values = less frequent sampling = fewer data points

- **`--duration`**: How long to keep collecting metrics before stopping
  - Accepts time units: `s` (seconds), `m` (minutes), `h` (hours)
  - Example: `--duration 1h` means run for 1 hour total
  - The tool will automatically stop after this time period

**How they work together:**

The number of samples collected = `duration / frequency`

**Examples:**

| Frequency | Duration | Total Samples | Use Case |
|-----------|----------|---------------|----------|
| 60s | 45m | 45 samples | Default - balanced monitoring |
| 30s | 1h | 120 samples | More frequent sampling for detailed analysis |
| 120s | 2h | 60 samples | Less frequent, longer observation period |
| 10s | 5m | 30 samples | Quick test or troubleshooting |
| 300s (5m) | 24h | 288 samples | Long-term monitoring with less granularity |

**Choosing the right values:**

- **Development/Testing**: Short duration (5-10m), frequent sampling (30-60s)
  ```bash
  --frequency 30 --duration 10m
  ```

- **Production Monitoring**: Longer duration (1-24h), moderate sampling (60-120s)
  ```bash
  --frequency 60 --duration 24h
  ```

- **Troubleshooting**: Very frequent sampling (10-30s), short duration (5-15m)
  ```bash
  --frequency 10 --duration 5m
  ```

## Database Support

The tool supports two database backends. **SQLite is used by default** when no `--db-type` flag is specified.

### SQLite (Default)
- **Default behavior** - no configuration needed
- Data stored in local file: `collected-data/kpi_metrics.db`
- Automatically created on first run
- No external dependencies

### PostgreSQL
- Requires `--db-type postgres` and `--postgres-url` flags
- Requires PostgreSQL server (version 9.5+)
- Connection URL formats:
  - **Standard URL**: `postgresql://user:password@host:port/dbname?sslmode=disable`
  - **Key-value**: `host=host port=port user=user password=pass dbname=dbname sslmode=disable`

# Local Grafana Setup for KPI Dashboard

This guide explains how developers can quickly run a local Grafana instance with the KPI dashboard pre-installed.

## Using Make

Run Grafana locally with one command:

make install-grafana

This will:

- Launch a Docker container named grafana-kpi
- Map local folders for provisioning Datasources and Dashboards
- Expose Grafana on port 3000

## Verify Grafana is running

docker ps | grep grafana-kpi

- Status should show Up
- Port mapping should show 0.0.0.0:3000->3000/tcp

## Open Grafana in browser

http://localhost:3000

- Default login: admin/admin
- Change password on first login

## Verify Datasource

1. Go to Settings → Data Sources
2. Select frser-sqlite-datasource
3. Click Test connection → should see "Data source is working"

## Verify KPI Dashboard

In disconnected environments, dashboards are automatically provisioned.

Ensure these files exist in the repository:

1. grafana/provisioning/dashboards/dashboard.yaml
2. grafana/dashboard/sqlite-dashboard.json
3. grafana/datasource/sqlite-datasource.yaml

- Run Grafana with:

  make install-grafana

- Open http://localhost:3000 and verify:
  1. Datasource frser-sqlite-datasource is listed under Configuration → Data Sources
  2. KPI dashboard appears under Dashboards → Manage and all graphs load

## Directory structure

kpi-collection-tool/
├── grafana/
│   ├── datasource/
│   │   └── sqlite-datasource.yaml
│   ├── dashboard/
│   │   └── sqlite-dashboard.json
│   └── provisioning/
│       └── dashboards/
│           └── dashboard.yaml
├── cmd/
│   └── kpi-collector/
│       └── main.go
└── Makefile

- grafana/datasource → Datasource YAML files
- grafana/dashboard → Dashboard JSON files
- Makefile → Automates running Grafana locally

