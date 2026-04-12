# kpi-collector AI Skill

An AI agent skill that teaches AI coding assistants how to use the `kpi-collector` CLI
and generate Telco-specific PromQL queries for OpenShift cluster monitoring.
Works with Cursor, Claude Code, and any IDE that supports the
[Agent Skills](https://agentskills.io/) open standard.

## What This Skill Does

When installed, your AI coding assistant will be able to:

- **Run kpi-collector commands** — collect metrics, query stored data, manage Grafana dashboards
- **Generate `kpis.json` files** — with battle-tested PromQL queries for Telco workloads
- **Cover Telco domains** — RAN/DU, PTP timing, SRIOV networking, 5G Core, cluster health
- **Troubleshoot issues** — knows common gotchas like Thanos staleness, TLS certs, and sandbox limitations


## Installation

### Quick Install (no clone needed)

If you installed kpi-collector via `go install` and don't have the repo locally,
download the skill files directly:

```bash
# Pick your IDE's skill directory
SKILL_DIR=~/.cursor/skills/kpi-collector   # Cursor
# SKILL_DIR=~/.claude/skills/kpi-collector  # Claude Code

mkdir -p "$SKILL_DIR"

REPO_URL="https://raw.githubusercontent.com/redhat-best-practices-for-k8s/kpi-collection-tool/main/docs/ai-skill"
curl -sL "$REPO_URL/SKILL.md" -o "$SKILL_DIR/SKILL.md"
curl -sL "$REPO_URL/telco-promql.md" -o "$SKILL_DIR/telco-promql.md"
```

### From a cloned repo

Both Cursor and Claude Code discover skills from the same directory structure:

| Location | Scope | When to use |
|----------|-------|-------------|
| Project: `.cursor/skills/` or `.claude/skills/` | This project only | You work with kpi-collector in this repo |
| Personal: `~/.cursor/skills/` or `~/.claude/skills/` | All your projects | You want the skill available everywhere |

Copy the skill files (pick the directory matching your IDE):

```bash
# Project-level skill
mkdir -p .cursor/skills/kpi-collector   # Cursor
mkdir -p .claude/skills/kpi-collector   # Claude Code
cp docs/ai-skill/SKILL.md docs/ai-skill/telco-promql.md <chosen-dir>/

# Personal skill (available across all projects)
mkdir -p ~/.cursor/skills/kpi-collector   # Cursor
mkdir -p ~/.claude/skills/kpi-collector   # Claude Code
cp docs/ai-skill/SKILL.md docs/ai-skill/telco-promql.md <chosen-dir>/
```

**Note**: `.cursor/` and `.claude/` are typically gitignored, so project-level skills
won't be committed. That's why the source files live in `docs/ai-skill/`.

## Usage Examples

Once installed, just ask your AI assistant naturally:

| What you say | What the skill helps with |
|---|---|
| "Check my cluster's health" | Generates a health-check `kpis.json`, asks for auth details, runs collection |
| "Create KPIs for a RAN cluster" | Asks which areas (CPU, PTP, networking), assembles a tailored `kpis.json` |
| "Generate PTP monitoring queries" | Produces PTP-specific KPIs (clock state, offset, clock class, delay) |
| "Collect CPU and memory every 30s for 2 hours" | Sets up `kpis.json` and runs with `--frequency 30s --duration 2h` |
| "Monitor my cluster overnight, sample every 5 minutes" | Configures a long-running collection with `--frequency 5m --duration 12h` |
| "Collect PTP offset once and network stats every minute" | Creates a mixed `kpis.json` with `run-once` for PTP and `sample-frequency` for networking |
| "Show me the last hour of CPU metrics for cluster prod-ran" | Runs `kpi-collector db show kpis` with `--since 1h --cluster-name prod-ran` |
| "Export all KPIs from today as CSV" | Runs `kpi-collector db show kpis --since 24h -o csv` |
| "Start Grafana to view results" | Runs `kpi-collector grafana start` with the right datasource |
| "My KPI query returns no data" | Walks through debugging: checks errors, suggests PromQL fixes |
| "Monitor 5G Core network functions" | Generates AMF/SMF/UPF resource KPIs with correct namespace filters |

## Skill Structure

```
docs/ai-skill/
├── SKILL.md          # Main skill — CLI reference, kpis.json format, common KPIs, gotchas, workflows
├── telco-promql.md   # Extended PromQL library — cluster health, node/pod resources,
│                     #   RAN/DU, PTP, networking, SRIOV, 5G Core, range query examples
└── README.md         # This file
```

## Requirements

- **kpi-collector** binary installed (`make install` or `make build`)
- **Access to an OpenShift cluster** via kubeconfig or bearer token + Thanos URL
- For Grafana: **Docker or Podman** installed and running
