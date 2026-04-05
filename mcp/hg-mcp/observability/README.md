# Observability Stack

Prometheus + Grafana + OpenTelemetry Collector for hg-mcp metrics and tracing.

## Quick Start

```bash
cd observability
docker compose up -d
```

## Dashboards

| Service | URL | Credentials |
|---------|-----|-------------|
| Grafana | http://localhost:3000 | admin / `$GRAFANA_ADMIN_PASSWORD` (default: `changeme`) |
| Prometheus | http://localhost:9090 | — |
| OTel Collector health | http://localhost:13133 | — |

## Ports

| Port | Service |
|------|---------|
| 3000 | Grafana |
| 9090 | Prometheus |
| 4317 | OTLP gRPC receiver |
| 4318 | OTLP HTTP receiver |
| 8888 | Collector self-metrics |
| 8889 | Prometheus exporter |

## Exported Metrics

hg-mcp exports metrics via mcpkit's observability middleware:

- `mcp_tool_invocations_total` — tool call count by name and status
- `mcp_tool_duration_seconds` — tool call duration histogram
- `mcp_tool_errors_total` — tool error count
- `mcp_tool_active_count` — currently executing tool calls

## Pre-provisioned Dashboards

- **hg-mcp** (`grafana/provisioning/dashboards/hg-mcp.json`) — tool invocations, latency, errors
- **Music Sync** (`grafana/provisioning/dashboards/music-sync.json`) — sync pipeline metrics
