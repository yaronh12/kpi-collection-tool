# KPI Collection Tool

CLI tool to automate KPI metrics collection from Prometheus/Thanos and visualize results in Grafana.

## Installation

Install directly from GitHub:

```bash
go install github.com/redhat-best-practices-for-k8s/kpi-collection-tool/cmd/kpi-collector@latest
```

If needed, add Go binaries to your PATH:

```bash
export PATH="$HOME/go/bin:$PATH"
```

For source-based installation and uninstall instructions, see [docs/installation.md](docs/installation.md).

## Quick Start (5 Minutes)

1. Generate a KPI configuration file for your cluster profile:

```bash
kpi-collector kpis generate ran
```

2. Collect KPI metrics:

```bash
kpi-collector run \
  --cluster-name my-cluster \
  --cluster-type ran \
  --kubeconfig ~/.kube/config \
  --kpis-file ran-kpis.json
```

3. Query collected data:

```bash
kpi-collector db show clusters
```

4. Visualize in Grafana:

```bash
kpi-collector grafana start --datasource=sqlite
```

## Command Map

- `kpi-collector kpis generate`: generate a kpis.json file for a cluster profile
- `kpi-collector run`: collect KPI metrics
- `kpi-collector db show`: query collected data
- `kpi-collector db remove`: remove stored data
- `kpi-collector grafana start|stop`: manage local Grafana dashboard

Get help anytime:

```bash
kpi-collector --help
kpi-collector run --help
kpi-collector db --help
kpi-collector grafana --help
```

## Documentation

- [Installation](docs/installation.md)
- [Collecting Metrics](docs/collecting-metrics.md)
- [KPI Profiles](kpi-profiles/README.md)
- [Database Commands](docs/database-commands.md)
- [Grafana](docs/grafana.md)
- [Troubleshooting](docs/troubleshooting.md)

