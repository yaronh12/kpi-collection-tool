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

The tool supports two authentication modes and two database backends (SQLite and PostgreSQL).

### Authentication Modes

#### 1. Using Kubeconfig (Automatic Discovery)

Automatically discovers Thanos URL and creates a service account token.

**Basic usage (uses SQLite by default):**
```bash
./kpi-collector --cluster-name my-cluster --kubeconfig ~/.kube/config
```

**With custom sampling parameters:**
```bash
./kpi-collector \
  --cluster-name my-cluster \
  --kubeconfig ~/.kube/config \
  --frequency 30 \
  --duration 1h \
  --output my-metrics.json \
  --log my-app.log
```

**Explicitly using SQLite:**
```bash
./kpi-collector \
  --cluster-name my-cluster \
  --kubeconfig ~/.kube/config \
  --db-type sqlite
```

**Using PostgreSQL:**
```bash
./kpi-collector \
  --cluster-name my-cluster \
  --kubeconfig ~/.kube/config \
  --db-type postgres \
  --postgres-url "postgresql://myuser:mypass@localhost:5432/kpi_metrics?sslmode=disable"
```

#### 2. Using Manual Credentials

Provide Thanos URL and bearer token directly.

**Basic usage (uses SQLite by default):**
```bash
./kpi-collector \
  --cluster-name my-cluster \
  --token YOUR_BEARER_TOKEN \
  --thanos-url thanos-querier.example.com
```

**With custom sampling parameters:**
```bash
./kpi-collector \
  --cluster-name my-cluster \
  --token YOUR_BEARER_TOKEN \
  --thanos-url thanos-querier.example.com \
  --frequency 120 \
  --duration 30m \
  --output results.json
```

**Using PostgreSQL:**
```bash
./kpi-collector \
  --cluster-name my-cluster \
  --token YOUR_BEARER_TOKEN \
  --thanos-url thanos-querier.example.com \
  --db-type postgres \
  --postgres-url "postgresql://myuser:mypass@localhost:5432/kpi_metrics?sslmode=disable"
```

### Complete Examples

**Development setup with SQLite (default):**
```bash
./kpi-collector \
  --cluster-name dev-cluster \
  --kubeconfig ~/.kube/config \
  --frequency 60 \
  --duration 1h \
  --insecure-tls
```

**Production setup with PostgreSQL:**
```bash
./kpi-collector \
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

## Command Line Flags

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--cluster-name` | Yes | - | Name of the cluster being monitored |
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

