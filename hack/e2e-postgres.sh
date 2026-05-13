#!/usr/bin/env bash
#
# End-to-end test: PostgreSQL + kpi-collector + Grafana
#
# Spins up a PostgreSQL container, runs kpi-collector against a local
# Prometheus/Thanos endpoint with PostgreSQL as the database backend,
# then launches Grafana pointed at that database.
#
# Prerequisites:
#   - docker or podman
#   - A kubeconfig with access to an OpenShift/K8s cluster
#   - kpi-collector binary (built via `make build`)
#
# Usage:
#   ./hack/e2e-postgres.sh                                    # defaults (~/.kube/config)
#   KUBECONFIG=/path/to/kubeconfig ./hack/e2e-postgres.sh     # custom kubeconfig
#   ./hack/e2e-postgres.sh --cleanup                          # tear everything down

set -euo pipefail

# ── Configuration (override via environment) ─────────────────────────
PG_CONTAINER="${PG_CONTAINER:-kpi-postgres}"
PG_PORT="${PG_PORT:-5433}"
PG_USER="${PG_USER:-kpi}"
PG_PASSWORD="${PG_PASSWORD:-kpi}"
PG_DB="${PG_DB:-kpi_metrics}"
PG_IMAGE="${PG_IMAGE:-postgres:16-alpine}"

KUBECONFIG="${KUBECONFIG:-${HOME}/.kube/config}"
CLUSTER_NAME="${CLUSTER_NAME:-e2e-test}"
CLUSTER_TYPE="${CLUSTER_TYPE:-ran}"
GRAFANA_PORT="${GRAFANA_PORT:-3000}"
INSECURE_TLS="${INSECURE_TLS:-true}"

PG_URL="postgresql://${PG_USER}:${PG_PASSWORD}@localhost:${PG_PORT}/${PG_DB}?sslmode=disable"
# Grafana runs in a container — it needs a host-reachable address, not localhost
PG_URL_GRAFANA="postgresql://${PG_USER}:${PG_PASSWORD}@host.docker.internal:${PG_PORT}/${PG_DB}?sslmode=disable"
KPI_COLLECTOR="./kpi-collector"
KPI_FILE="${KPI_FILE:-hack/kpis-e2e-postgres.yaml}"

# ── Detect container runtime ─────────────────────────────────────────
detect_runtime() {
    if command -v podman &>/dev/null && podman info &>/dev/null; then
        echo "podman"
    elif command -v docker &>/dev/null && docker info &>/dev/null; then
        echo "docker"
    else
        echo "ERROR: no working container runtime found (tried podman, docker)" >&2
        exit 1
    fi
}

RUNTIME="$(detect_runtime)"
echo "Container runtime: ${RUNTIME}"

# ── Cleanup mode ─────────────────────────────────────────────────────
cleanup() {
    echo ""
    echo "=== Cleaning up ==="
    ${RUNTIME} stop "${PG_CONTAINER}" 2>/dev/null || true
    ${RUNTIME} rm -f "${PG_CONTAINER}" 2>/dev/null || true
    ${KPI_COLLECTOR} grafana stop 2>/dev/null || true
    echo "Done."
}

if [[ "${1:-}" == "--cleanup" ]]; then
    cleanup
    exit 0
fi

trap cleanup EXIT

# ── Pre-flight checks ────────────────────────────────────────────────
if [[ ! -x "${KPI_COLLECTOR}" ]]; then
    echo "Binary not found at ${KPI_COLLECTOR} — building..."
    make build
fi

if [[ ! -f "${KUBECONFIG}" ]]; then
    echo "ERROR: kubeconfig not found: ${KUBECONFIG}" >&2
    exit 1
fi

if [[ ! -f "${KPI_FILE}" ]]; then
    echo "ERROR: KPI file not found: ${KPI_FILE}" >&2
    exit 1
fi

# ── Step 1: Start PostgreSQL ─────────────────────────────────────────
echo ""
echo "=== Step 1: Starting PostgreSQL (${PG_CONTAINER}) ==="
${RUNTIME} stop "${PG_CONTAINER}" 2>/dev/null || true
${RUNTIME} rm -f "${PG_CONTAINER}" 2>/dev/null || true

${RUNTIME} run -d \
    --name "${PG_CONTAINER}" \
    -p "${PG_PORT}:5432" \
    -e "POSTGRES_USER=${PG_USER}" \
    -e "POSTGRES_PASSWORD=${PG_PASSWORD}" \
    -e "POSTGRES_DB=${PG_DB}" \
    "${PG_IMAGE}"

echo "Waiting for PostgreSQL to be ready..."
for i in $(seq 1 30); do
    if ${RUNTIME} exec "${PG_CONTAINER}" pg_isready -U "${PG_USER}" &>/dev/null; then
        echo "PostgreSQL is ready (attempt ${i})"
        break
    fi
    if [[ "${i}" -eq 30 ]]; then
        echo "ERROR: PostgreSQL failed to start within 30 seconds" >&2
        exit 1
    fi
    sleep 1
done

# ── Step 2: Collect KPIs ─────────────────────────────────────────────
echo ""
echo "=== Step 2: Collecting KPIs ==="
echo "  Kubeconfig : ${KUBECONFIG}"
echo "  Cluster    : ${CLUSTER_NAME} (${CLUSTER_TYPE})"
echo "  Database   : PostgreSQL @ localhost:${PG_PORT}/${PG_DB}"
echo "  KPI file   : ${KPI_FILE}"
echo ""

RUN_ARGS=(
    --cluster-name "${CLUSTER_NAME}"
    --cluster-type "${CLUSTER_TYPE}"
    --kubeconfig "${KUBECONFIG}"
    --kpis-file "${KPI_FILE}"
    --db-type postgres
    --postgres-url "${PG_URL}"
    --once
)

if [[ "${INSECURE_TLS}" == "true" ]]; then
    RUN_ARGS+=(--insecure-tls)
fi

${KPI_COLLECTOR} run "${RUN_ARGS[@]}"

# ── Step 3: Verify data ──────────────────────────────────────────────
echo ""
echo "=== Step 3: Verifying stored data ==="

echo "--- Clusters ---"
${KPI_COLLECTOR} db show clusters \
    --db-type postgres --postgres-url "${PG_URL}"

echo ""
echo "--- KPIs ---"
${KPI_COLLECTOR} db show kpis \
    --db-type postgres --postgres-url "${PG_URL}"

echo ""
echo "--- Errors ---"
${KPI_COLLECTOR} db show errors \
    --db-type postgres --postgres-url "${PG_URL}"

# ── Step 4: Start Grafana ────────────────────────────────────────────
echo ""
echo "=== Step 4: Starting Grafana ==="
${KPI_COLLECTOR} grafana start \
    --datasource=postgres \
    --postgres-url "${PG_URL_GRAFANA}" \
    --port "${GRAFANA_PORT}"

# ── Done ─────────────────────────────────────────────────────────────
echo ""
echo "============================================"
echo "  E2E setup complete!"
echo ""
echo "  PostgreSQL : localhost:${PG_PORT} (${PG_USER}/${PG_PASSWORD})"
echo "  Grafana    : http://localhost:${GRAFANA_PORT} (admin/admin)"
echo ""
echo "  To tear down:  ./hack/e2e-postgres.sh --cleanup"
echo "============================================"

trap - EXIT
