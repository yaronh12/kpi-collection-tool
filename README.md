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

1. Collect KPI metrics:

```bash
kpi-collector run \
  --cluster-name my-cluster \
  --cluster-type ran \
  --kubeconfig ~/.kube/config \
  --kpis-file kpis.json
```

2. Query collected data:

```bash
kpi-collector db show clusters
```

3. Visualize in Grafana:

```bash
kpi-collector grafana start --datasource=sqlite
```

## Command Map

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

## AI Skill for Cursor / Claude Code

An AI agent skill is included that teaches your coding assistant how to use
kpi-collector and generate Telco-specific PromQL queries.
See [docs/ai-skill/](docs/ai-skill/) for installation and usage instructions.

## Documentation

- [Documentation Index](docs/index.md)
- [Installation](docs/installation.md)
- [Collecting Metrics](docs/collecting-metrics.md)
- [Database Commands](docs/database-commands.md)
- [Grafana](docs/grafana.md)
- [Troubleshooting](docs/troubleshooting.md)

