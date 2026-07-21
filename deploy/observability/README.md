# Pole 可观测性 Docker Compose 兼容栈

本目录保留不具备 Kubernetes 时使用的 GreptimeDB + OpenTelemetry Collector quickstart。当前默认本地交付方式是 [`deploy/kubernetes`](../kubernetes/README.md)，它会把 Pole Control Plane、GreptimeDB 与 Collector 部署到同一个 `pole-system` namespace，并继续访问宿主机 MySQL `3306`。

```bash
cd deploy/observability
docker compose up -d
```

暴露端口：

| 组件 | 端口 | 用途 |
|---|---:|---|
| OpenTelemetry Collector | `4317` | OTLP/gRPC receiver |
| OpenTelemetry Collector | `4318` | OTLP/HTTP receiver |
| OpenTelemetry Collector | `13133` | Collector health check |
| OpenTelemetry Collector | `8888` | Collector 自身 Prometheus metrics |
| GreptimeDB | `4000` | HTTP / OTLP HTTP |
| GreptimeDB | `14001` | gRPC，映射到容器内 `4001`，可通过 `GREPTIMEDB_GRPC_PORT` 覆盖 |
| GreptimeDB | `4002` | MySQL 协议 |
| GreptimeDB | `4003` | PostgreSQL 协议 |

Collector 接收 SDK、sidecar 和 control-plane 的 OTLP signals，并将 traces、metrics、logs 分别写入 GreptimeDB：

- traces: `otlp_http/greptime_traces`，使用 `greptime_trace_v1` pipeline。
- metrics: `otlp_http/greptime_metrics`。
- logs: `otlp_http/greptime_logs`，默认写入 `pole_events` 表；本平台只允许结构化 event 和 audit 走 OTel logs，不采集业务普通日志。

验证 Collector 与 GreptimeDB 写入链路：

```bash
NOW_NS=$(date +%s%N)
curl -fsS -X POST http://127.0.0.1:4318/v1/logs \
  -H 'Content-Type: application/json' \
  -d "{\"resourceLogs\":[{\"resource\":{\"attributes\":[{\"key\":\"service.name\",\"value\":{\"stringValue\":\"pole-smoke\"}},{\"key\":\"pole.namespace\",\"value\":{\"stringValue\":\"default\"}},{\"key\":\"pole.service.name\",\"value\":{\"stringValue\":\"pole-smoke\"}}]},\"scopeLogs\":[{\"scope\":{\"name\":\"pole-smoke\"},\"logRecords\":[{\"timeUnixNano\":\"${NOW_NS}\",\"severityText\":\"INFO\",\"body\":{\"stringValue\":\"pole observability smoke event\"},\"attributes\":[{\"key\":\"event.name\",\"value\":{\"stringValue\":\"pole.smoke.event\"}},{\"key\":\"pole.result\",\"value\":{\"stringValue\":\"success\"}}]}]}]}]}"

curl -fsS -X POST \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'sql=select timestamp, severity_text, body, log_attributes, resource_attributes from pole_events order by timestamp desc limit 1' \
  'http://127.0.0.1:4000/v1/sql?db=public'
```

停止本地栈：

```bash
cd deploy/observability
docker compose down
```

清理 GreptimeDB 本地数据：

```bash
cd deploy/observability
docker compose down -v
```
