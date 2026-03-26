# Troubleshooting

## "No data" in dashboard

1. Ensure data was collected first with `kpi-collector run`
2. Check Grafana time range (top-right), for example "Last 24 hours" or "Last 7 days"
3. Verify the KPI dropdown has a selected value
4. For SQLite, ensure `./kpi-collector-artifacts/kpi_metrics.db` exists in the directory where you ran `kpi-collector run`
5. For PostgreSQL, test datasource connectivity in Grafana (**Settings** -> **Data Sources**)

## PostgreSQL connection errors

1. Verify PostgreSQL is running: `psql -l`
2. Check the connection URL
3. For local PostgreSQL with Docker, use:
   - `host.docker.internal` on Mac/Windows
   - `172.17.0.1` on Linux
4. Test connection directly: `psql "your-connection-url"`
