# Collecting Metrics

The `run` command gathers KPI metrics from Prometheus/Thanos and stores them in a database.

Related guides:
- [Installation](installation.md)
- [Database Commands](database-commands.md)
- [Grafana](grafana.md)

## Authentication Modes

### 1) Using Kubeconfig (Automatic Discovery)

Automatically discovers Thanos URL and creates a service account token.

Basic usage (SQLite by default):

```bash
kpi-collector run \
  --cluster-name my-cluster \
  --cluster-type ran \
  --kubeconfig ~/.kube/config \
  --kpis-file kpis.json
```

With custom sampling parameters:

```bash
kpi-collector run \
  --cluster-name my-cluster \
  --cluster-type ran \
  --kubeconfig ~/.kube/config \
  --kpis-file kpis.json \
  --frequency 30s \
  --duration 1h
```

Explicitly using SQLite:

```bash
kpi-collector run \
  --cluster-name my-cluster \
  --kubeconfig ~/.kube/config \
  --kpis-file kpis.json \
  --db-type sqlite
```

Using PostgreSQL:

```bash
kpi-collector run \
  --cluster-name my-cluster \
  --kubeconfig ~/.kube/config \
  --kpis-file kpis.json \
  --db-type postgres \
  --postgres-url "postgresql://myuser:mypass@localhost:5432/kpi_metrics?sslmode=disable"
```

### 2) Using Token and Thanos URL

Provide Thanos URL and bearer token directly.

#### Obtaining Thanos URL and Bearer Token from OpenShift

```bash
# Get the Thanos querier URL from the route in your monitoring namespace
export THANOS_URL=$(oc get route <thanos-route-name> -n <monitoring-namespace> -o jsonpath='{.spec.host}')

# Create a bearer token using a service account with access to Thanos
export TOKEN=$(oc create token <service-account-name> -n <monitoring-namespace> --duration=<duration>)
```

Common values include:
- Namespace: `openshift-monitoring`
- Thanos route: `thanos-querier`
- Service account: a service account with permissions to query metrics

Then pass these values with `--token` and `--thanos-url`:

```bash
kpi-collector run \
  --cluster-name my-cluster \
  --token $TOKEN \
  --thanos-url $THANOS_URL \
  --kpis-file kpis.json
```

## `--insecure-tls`

Use this flag when running against clusters or Prometheus/Thanos servers with self-signed or untrusted certificates.

```bash
kpi-collector run \
  --cluster-name my-cluster \
  --kubeconfig ~/.kube/config \
  --kpis-file kpis.json \
  --insecure-tls
```

What it does:
- Skips TLS certificate verification for HTTPS requests
- Applies to Kubernetes API calls and Thanos/Prometheus queries
- Helps in environments where the certificate cannot be validated

When to use:
- Self-signed certificates
- Disconnected, lab, or air-gapped clusters
- kubeconfig without a valid CA
- Errors such as `x509: certificate signed by unknown authority`

Complete example:

```bash
kpi-collector run \
  --cluster-name dev-cluster \
  --kubeconfig ~/.kube/config \
  --kpis-file kpis.json \
  --frequency 60 \
  --duration 1h \
  --insecure-tls
```

## Command Line Flags (`run`)

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--cluster-name` | Yes | - | Name of the cluster being monitored |
| `--cluster-type` | Yes | - | Cluster type for categorization: `ran`, `core`, or `hub` |
| `--kubeconfig` | No* | - | Path to kubeconfig file for auto-discovery |
| `--token` | No* | - | Bearer token for Thanos authentication |
| `--thanos-url` | No* | - | Thanos querier URL (without `https://`) |
| `--insecure-tls` | No | false | Skip TLS certificate verification (dev only) |
| `--frequency` | No | 1m | Sampling frequency (for example: `10s`, `1m`, `2h`, `24h`) |
| `--duration` | No | 45m | Total sampling duration (for example: `10s`, `1m`, `2h`, `24h`) |
| `--db-type` | No | sqlite | Database type: `sqlite` or `postgres` |
| `--postgres-url` | No** | - | PostgreSQL connection string |
| `--once` | No | false | Collect all KPIs once and exit (ignores `--frequency` and `--duration`) |
| `--kpis-file` | Yes | - | Path to KPIs configuration file (see `kpis.json.template`) |

\* Either provide `--kubeconfig` OR both `--token` and `--thanos-url`  
\*\* Required when `--db-type=postgres`

## Dynamic CPU Placeholders

Queries can use `{{RESERVED_CPUS}}` and `{{ISOLATED_CPUS}}` placeholders. These are replaced with CPU IDs fetched from `PerformanceProfile` CRs in the cluster (for example, `"0-1"` becomes `"0|1"`). This feature requires `--kubeconfig` authentication. See `kpis.json.template` for examples.

## Understanding Frequency and Duration

The `--frequency` and `--duration` flags work together to control how metrics are collected.

- `--frequency`: how often to collect metrics
  - Example: `--frequency 60` means once per minute
  - Lower values collect more data points
  - Higher values collect fewer data points
- `--duration`: total collection time before stopping
  - Accepts `s`, `m`, `h`
  - Example: `--duration 1h`

Number of samples collected:

`duration / frequency`

Examples:

| Frequency | Duration | Total Samples | Use Case |
|-----------|----------|---------------|----------|
| 60s | 45m | 45 samples | Default, balanced monitoring |
| 30s | 1h | 120 samples | Detailed analysis |
| 120s | 2h | 60 samples | Longer observation, less granularity |
| 10s | 5m | 30 samples | Quick troubleshooting |
| 300s (5m) | 24h | 288 samples | Long-term monitoring |

Choosing values:

- Development/testing:
  ```bash
  --frequency 30 --duration 10m
  ```
- Production monitoring:
  ```bash
  --frequency 60 --duration 24h
  ```
- Troubleshooting:
  ```bash
  --frequency 10 --duration 5m
  ```

## Single Run Mode

Use `--once` to collect every KPI metric exactly once and exit immediately. When this flag is set, `--frequency` and `--duration` are ignored.

This is useful for:
- One-off snapshots of cluster metrics
- CI/CD pipelines that need a single data point
- Quick validation that queries and connectivity work
- Range queries (e.g. `rate(...[5m])`, `avg_over_time(...[1h])`) that already aggregate over a time window and don't need repeated sampling

```bash
kpi-collector run \
  --cluster-name my-cluster \
  --cluster-type ran \
  --kubeconfig ~/.kube/config \
  --kpis-file kpis.json \
  --once
```

With manual credentials:

```bash
kpi-collector run \
  --cluster-name my-cluster \
  --cluster-type core \
  --token $TOKEN \
  --thanos-url $THANOS_URL \
  --kpis-file kpis.json \
  --once
```

## Per-Query `run-once`

Individual KPI queries can be marked with `"run-once": true` in the KPI configuration file. These queries execute once at the start of collection and are excluded from the repeated sampling loop, even when running with `--frequency` and `--duration`.

This is useful for queries that only need a single data point, such as:
- Range queries that already aggregate over a time window (e.g. `rate(...[30m])`)
- Static cluster information (uptime, version, node count)
- Baseline snapshots taken before continuous monitoring begins

Example configuration:

```json
{
    "kpis": [
        {
            "id": "cluster-uptime",
            "promquery": "max(time() - process_start_time_seconds{job=\"kubelet\"})",
            "run-once": true
        },
        {
            "id": "node-cpu-usage",
            "promquery": "avg by (instance) (rate(node_cpu_seconds_total{mode!=\"idle\"}[5m]))"
        }
    ]
}
```

In this example, `cluster-uptime` runs once immediately, while `node-cpu-usage` is sampled repeatedly at the configured frequency for the full duration.

If all queries in the configuration are marked `"run-once": true`, the collector executes them all once and exits without waiting for the duration timer.
