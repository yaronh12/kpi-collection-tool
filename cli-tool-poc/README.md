# RDS KPI Collection CLI Tool - POC

**Proof of Concept** for a Go-based command-line tool that will automate the collection of Red Hat Data Services (RDS) Key Performance Indicators (KPIs) by executing predefined OpenShift commands and Prometheus queries.

## Overview

This POC explores the feasibility of building a CLI tool to streamline RDS metrics collection:
- Execute a predefined list of `oc` (OpenShift CLI) commands
- Run Prometheus queries to collect performance metrics  
- Output structured JSON data ready for analysis and visualization

## Planned Features

- **Automated Data Collection**: Execute predefined commands without manual intervention
- **Structured Output**: Format all data as JSON for easy parsing and analysis
- **OpenShift Integration**: Leverage `oc` commands for cluster information
- **Prometheus Metrics**: Query Prometheus for performance and health metrics

## Proposed Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   CLI Tool      │    │   OpenShift      │    │   Prometheus    │
│                 │───▶│   Cluster        │    │   Server        │
│ - Command Exec  │    │                  │    │                 │
│ - Query Engine  │    │ - oc commands    │    │ - Metrics API   │
│ - JSON Output   │    │ - Resource info  │    │ - Time series   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## POC Goals

- Validate the approach for automated KPI collection
- Define the structure for commands and queries
- Design the JSON output format
- Identify technical challenges and requirements

