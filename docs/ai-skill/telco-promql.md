# Telco PromQL Reference for kpi-collector

Comprehensive PromQL queries for OpenShift Telco clusters. All queries are formatted
as kpis.json entries ready to copy.

## Cluster Health

```json
{
    "id": "apiserver-request-latency-p99",
    "promquery": "histogram_quantile(0.99, sum by (le, verb) (rate(apiserver_request_duration_seconds_bucket{job=\"apiserver\"}[5m])))"
},
{
    "id": "apiserver-request-rate",
    "promquery": "sum by (code) (rate(apiserver_request_total{job=\"apiserver\"}[5m]))"
},
{
    "id": "apiserver-error-rate",
    "promquery": "sum(rate(apiserver_request_total{job=\"apiserver\", code=~\"5..\"}[5m])) / sum(rate(apiserver_request_total{job=\"apiserver\"}[5m]))"
},
{
    "id": "etcd-leader-changes",
    "promquery": "changes(etcd_server_leader_changes_seen_total[1h])",
    "run-once": true
},
{
    "id": "etcd-db-size",
    "promquery": "etcd_mvcc_db_total_size_in_bytes"
},
{
    "id": "etcd-disk-wal-fsync-p99",
    "promquery": "histogram_quantile(0.99, rate(etcd_disk_wal_fsync_duration_seconds_bucket[5m]))"
},
{
    "id": "etcd-disk-backend-commit-p99",
    "promquery": "histogram_quantile(0.99, rate(etcd_disk_backend_commit_duration_seconds_bucket[5m]))"
},
{
    "id": "kubelet-running-pods",
    "promquery": "sum by (instance) (kubelet_running_pods)"
},
{
    "id": "kubelet-running-containers",
    "promquery": "sum by (instance) (kubelet_running_containers)"
},
{
    "id": "cluster-node-status",
    "promquery": "kube_node_status_condition{condition=\"Ready\", status=\"true\"}",
    "run-once": true
},
{
    "id": "cluster-uptime",
    "promquery": "max(time() - process_start_time_seconds{job=\"kubelet\"})",
    "run-once": true
}
```

## Node Resources

```json
{
    "id": "node-cpu-by-mode",
    "promquery": "avg by (instance, mode) (rate(node_cpu_seconds_total[5m]))"
},
{
    "id": "node-cpu-saturation",
    "promquery": "node_load1 / count without (cpu, mode) (node_cpu_seconds_total{mode=\"idle\"})"
},
{
    "id": "node-memory-total",
    "promquery": "node_memory_MemTotal_bytes",
    "run-once": true
},
{
    "id": "node-memory-available-percent",
    "promquery": "100 * node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes"
},
{
    "id": "node-memory-hugepages-free",
    "promquery": "node_memory_HugePages_Free"
},
{
    "id": "node-memory-hugepages-total",
    "promquery": "node_memory_HugePages_Total",
    "run-once": true
},
{
    "id": "node-disk-io-read",
    "promquery": "rate(node_disk_read_bytes_total[5m])"
},
{
    "id": "node-disk-io-write",
    "promquery": "rate(node_disk_written_bytes_total[5m])"
},
{
    "id": "node-disk-iops",
    "promquery": "rate(node_disk_io_time_seconds_total[5m])"
},
{
    "id": "node-filesystem-usage",
    "promquery": "100 - (node_filesystem_avail_bytes{mountpoint=\"/\"} / node_filesystem_size_bytes{mountpoint=\"/\"} * 100)"
}
```

## Pod / Container Resources

```json
{
    "id": "pod-cpu-usage",
    "promquery": "sort_desc(sum by (pod, namespace) (rate(container_cpu_usage_seconds_total{container!=\"\"}[5m])))"
},
{
    "id": "pod-cpu-throttling",
    "promquery": "sum by (pod, namespace) (rate(container_cpu_cfs_throttled_periods_total[5m])) / sum by (pod, namespace) (rate(container_cpu_cfs_periods_total[5m])) > 0"
},
{
    "id": "pod-memory-working-set",
    "promquery": "sort_desc(sum by (pod, namespace) (container_memory_working_set_bytes{container!=\"\"}))"
},
{
    "id": "pod-memory-rss",
    "promquery": "sort_desc(sum by (pod, namespace) (container_memory_rss{container!=\"\"}))"
},
{
    "id": "pod-oom-kills",
    "promquery": "kube_pod_container_status_last_terminated_reason{reason=\"OOMKilled\"}"
},
{
    "id": "pod-restart-rate",
    "promquery": "rate(kube_pod_container_status_restarts_total[1h]) > 0"
},
{
    "id": "pod-not-ready",
    "promquery": "kube_pod_status_ready{condition=\"false\"}"
}
```

## RAN / DU Specific

```json
{
    "id": "system-slice-cpu-all",
    "promquery": "sort_desc(rate(container_cpu_usage_seconds_total{id=~\"/system.slice/.*\"}[5m]))"
},
{
    "id": "ovs-slice-cpu",
    "promquery": "sort_desc(rate(container_cpu_usage_seconds_total{id=~\"/ovs.slice/.*\"}[5m]))"
},
{
    "id": "cpu-reserved-by-mode",
    "promquery": "sum by (cpu, mode) (rate(node_cpu_seconds_total{cpu=~\"{{RESERVED_CPUS}}\"}[5m]))"
},
{
    "id": "cpu-isolated-idle",
    "promquery": "rate(node_cpu_seconds_total{cpu=~\"{{ISOLATED_CPUS}}\", mode=\"idle\"}[5m])"
},
{
    "id": "cpu-isolated-nonIdle",
    "promquery": "1 - rate(node_cpu_seconds_total{cpu=~\"{{ISOLATED_CPUS}}\", mode=\"idle\"}[5m])"
},
{
    "id": "irq-by-cpu",
    "promquery": "rate(node_cpu_seconds_total{mode=\"irq\"}[5m])"
},
{
    "id": "softirq-by-cpu",
    "promquery": "rate(node_cpu_seconds_total{mode=\"softirq\"}[5m])"
},
{
    "id": "context-switches",
    "promquery": "rate(node_context_switches_total[5m])"
},
{
    "id": "du-pod-cpu",
    "promquery": "sum by (pod) (rate(container_cpu_usage_seconds_total{namespace=~\".*du.*\", container!=\"\"}[5m]))"
},
{
    "id": "du-pod-memory",
    "promquery": "sum by (pod) (container_memory_working_set_bytes{namespace=~\".*du.*\", container!=\"\"})"
}
```

## PTP / Timing (Extended)

```json
{
    "id": "ptp-offset-by-node",
    "promquery": "openshift_ptp_offset_ns * on(node) group_left() kube_node_info"
},
{
    "id": "ptp-frequency-adjustment",
    "promquery": "openshift_ptp_frequency_adjustment_ns"
},
{
    "id": "ptp-port-state",
    "promquery": "openshift_ptp_interface_role",
    "run-once": true
},
{
    "id": "ptp-process-restarts",
    "promquery": "changes(openshift_ptp_process_restart_count[1h])",
    "run-once": true
},
{
    "id": "ptp-offset-range-1h",
    "promquery": "abs(openshift_ptp_offset_ns)",
    "query-type": "range",
    "step": "10s",
    "range": "1h",
    "run-once": true
},
{
    "id": "ptp-gm-clock-class",
    "promquery": "openshift_ptp_clock_class{process=\"ts2phc\"}"
}
```

### PTP clock state values

| Value | Meaning |
|-------|---------|
| 0 | FREERUN — not synchronized |
| 1 | LOCKED — synchronized to GM |
| 2 | HOLDOVER — lost GM, using local oscillator |

### PTP clock class values

| Class | Meaning |
|-------|---------|
| 6 | Locked to primary reference (GNSS) |
| 7 | Previously locked, now holdover |
| 52 | Degraded, alternative holdover |
| 140 | Holdover, out of spec |
| 248 | Free-running |

## Networking (Extended)

```json
{
    "id": "interface-bandwidth-rx",
    "promquery": "rate(node_network_receive_bytes_total{device!~\"lo|veth.*|br.*\"}[5m]) * 8"
},
{
    "id": "interface-bandwidth-tx",
    "promquery": "rate(node_network_transmit_bytes_total{device!~\"lo|veth.*|br.*\"}[5m]) * 8"
},
{
    "id": "interface-packet-rate-rx",
    "promquery": "rate(node_network_receive_packets_total{device!~\"lo|veth.*|br.*\"}[5m])"
},
{
    "id": "interface-packet-rate-tx",
    "promquery": "rate(node_network_transmit_packets_total{device!~\"lo|veth.*|br.*\"}[5m])"
},
{
    "id": "interface-drop-rate",
    "promquery": "rate(node_network_receive_drop_total[5m]) + rate(node_network_transmit_drop_total[5m])"
},
{
    "id": "interface-error-rate",
    "promquery": "rate(node_network_receive_errs_total[5m]) + rate(node_network_transmit_errs_total[5m])"
},
{
    "id": "container-rx-all-interfaces",
    "promquery": "sort_desc(sum by (pod, interface) (rate(container_network_receive_bytes_total[5m])))"
},
{
    "id": "container-tx-all-interfaces",
    "promquery": "sort_desc(sum by (pod, interface) (rate(container_network_transmit_bytes_total[5m])))"
}
```

### SRIOV (if sriov-network-metrics-exporter is deployed)

```json
{
    "id": "sriov-vf-rx-bytes",
    "promquery": "rate(sriov_vf_rx_bytes[5m])"
},
{
    "id": "sriov-vf-tx-bytes",
    "promquery": "rate(sriov_vf_tx_bytes[5m])"
},
{
    "id": "sriov-vf-rx-packets",
    "promquery": "rate(sriov_vf_rx_packets[5m])"
},
{
    "id": "sriov-vf-tx-packets",
    "promquery": "rate(sriov_vf_tx_packets[5m])"
},
{
    "id": "sriov-vf-rx-dropped",
    "promquery": "rate(sriov_vf_rx_dropped[5m]) > 0"
},
{
    "id": "sriov-vf-tx-dropped",
    "promquery": "rate(sriov_vf_tx_dropped[5m]) > 0"
}
```

## 5G Core Network Functions

Adapt namespace filters to match the deployment (e.g. `open5gs`, `free5gc`, vendor-specific).

```json
{
    "id": "amf-cpu",
    "promquery": "sum by (pod) (rate(container_cpu_usage_seconds_total{namespace=~\".*5gc.*\", pod=~\".*amf.*\", container!=\"\"}[5m]))"
},
{
    "id": "smf-cpu",
    "promquery": "sum by (pod) (rate(container_cpu_usage_seconds_total{namespace=~\".*5gc.*\", pod=~\".*smf.*\", container!=\"\"}[5m]))"
},
{
    "id": "upf-cpu",
    "promquery": "sum by (pod) (rate(container_cpu_usage_seconds_total{namespace=~\".*5gc.*\", pod=~\".*upf.*\", container!=\"\"}[5m]))"
},
{
    "id": "amf-memory",
    "promquery": "sum by (pod) (container_memory_working_set_bytes{namespace=~\".*5gc.*\", pod=~\".*amf.*\", container!=\"\"})"
},
{
    "id": "smf-memory",
    "promquery": "sum by (pod) (container_memory_working_set_bytes{namespace=~\".*5gc.*\", pod=~\".*smf.*\", container!=\"\"})"
},
{
    "id": "upf-memory",
    "promquery": "sum by (pod) (container_memory_working_set_bytes{namespace=~\".*5gc.*\", pod=~\".*upf.*\", container!=\"\"})"
}
```

## Range Query Examples

Use range queries when you need historical data with specific resolution:

```json
{
    "id": "cpu-trend-1h",
    "promquery": "avg by (instance) (rate(node_cpu_seconds_total{mode!=\"idle\"}[5m]))",
    "query-type": "range",
    "step": "30s",
    "range": "1h",
    "run-once": true
},
{
    "id": "memory-trend-6h",
    "promquery": "node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes",
    "query-type": "range",
    "step": "1m",
    "range": "6h",
    "sample-frequency": "6h"
},
{
    "id": "ptp-offset-trend-24h",
    "promquery": "abs(openshift_ptp_offset_ns)",
    "query-type": "range",
    "step": "30s",
    "range": "24h",
    "run-once": true
}
```

## PromQL Tips for Thanos

1. **Use wider rate windows** — `[5m]` minimum. Thanos deduplication can cause gaps
   with `[1m]` or `[2m]` windows.

2. **Prefer `avg_over_time` for smoothing** — when querying through Thanos,
   `avg_over_time(metric[5m])` produces smoother results than raw instant queries.

3. **Label filtering matters** — always filter `container!=""` in container metrics
   to exclude the pause container aggregation.

4. **`topk` and `sort_desc`** — use `topk(N, ...)` to limit cardinality for
   high-cardinality metrics. Use `sort_desc(...)` when you want all values ordered.

5. **Regex for multi-value labels** — use `cpu=~"0|1|2|3"` for specific CPUs.
   The `{{RESERVED_CPUS}}` placeholder does this automatically.

6. **Absent metrics** — use `absent(metric_name)` to detect when an exporter is
   down or a metric is not being scraped.
