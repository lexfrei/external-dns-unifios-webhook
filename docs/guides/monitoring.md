# Monitoring

Set up monitoring and alerting for external-dns-unifios-webhook.

## Prometheus Metrics

The webhook exposes Prometheus metrics on the health server port.

### Metrics Endpoint

```text
http://<pod-ip>:8080/metrics
```

### Available Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `external_dns_unifi_dns_records_managed` | Gauge | Number of DNS records managed by type |
| `external_dns_unifi_dns_operations_total` | Counter | Total DNS operations (labels: operation, status) |
| `external_dns_unifi_dns_operation_duration_seconds` | Histogram | DNS operation latency |
| `external_dns_unifi_dns_changes_applied` | Histogram | Changes applied per batch (labels: change_type) |
| `external_dns_unifi_readiness_cache_hits_total` | Counter | Readiness cache hits |
| `external_dns_unifi_readiness_cache_misses_total` | Counter | Readiness cache misses |
| `external_dns_unifi_readiness_cache_age_seconds` | Gauge | Readiness cache age |

### Scrape Configuration

For Prometheus Operator with ServiceMonitor:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: external-dns-unifi
  namespace: external-dns-unifi
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: external-dns
  endpoints:
    - port: health
      path: /metrics
      interval: 30s
```

For standard Prometheus scrape config:

```yaml
scrape_configs:
  - job_name: 'external-dns-unifi'
    kubernetes_sd_configs:
      - role: pod
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_label_app_kubernetes_io_name]
        action: keep
        regex: external-dns
      - source_labels: [__meta_kubernetes_pod_container_port_name]
        action: keep
        regex: health
```

## Grafana Dashboard

### Key Panels

1. **Records Managed**: Current count of DNS records

    ```promql
    external_dns_unifi_dns_records_managed
    ```

2. **Operations Rate**: Operations per second

    ```promql
    rate(external_dns_unifi_dns_operations_total[5m])
    ```

3. **Operation Latency**: P99 latency

    ```promql
    histogram_quantile(0.99, rate(external_dns_unifi_dns_operation_duration_seconds_bucket[5m]))
    ```

4. **Error Rate**: Failed operations

    ```promql
    rate(external_dns_unifi_dns_operations_total{status="error"}[5m])
    ```

## Alerting

### PrometheusRule Examples

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: external-dns-unifi-alerts
  namespace: external-dns-unifi
spec:
  groups:
    - name: external-dns-unifi
      rules:
        - alert: ExternalDNSUniFiDown
          expr: up{job="external-dns-unifi"} == 0
          for: 5m
          labels:
            severity: critical
          annotations:
            summary: "external-dns-unifios-webhook is down"
            description: "The webhook has been down for more than 5 minutes."

        - alert: ExternalDNSUniFiHighErrorRate
          expr: |
            rate(external_dns_unifi_dns_operations_total{status="error"}[5m]) /
            rate(external_dns_unifi_dns_operations_total[5m]) > 0.1
          for: 10m
          labels:
            severity: warning
          annotations:
            summary: "High DNS operation error rate"
            description: "More than 10% of DNS operations are failing."

        - alert: ExternalDNSUniFiSlowOperations
          expr: |
            histogram_quantile(0.99, rate(external_dns_unifi_dns_operation_duration_seconds_bucket[5m])) > 5
          for: 10m
          labels:
            severity: warning
          annotations:
            summary: "DNS operations are slow"
            description: "P99 latency is above 5 seconds."
```

## Health Endpoints

### Liveness

```text
GET /healthz
```

Returns 200 if the webhook process is alive.

### Readiness

```text
GET /readyz
```

Returns 200 if the webhook can connect to UniFi controller and is ready to serve requests.

### Kubernetes Probes

```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 10
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /readyz
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 10
  failureThreshold: 3
```

## Logging

### Structured Logging

The webhook outputs JSON logs by default:

```json
{
  "time": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "msg": "Creating DNS record",
  "record": "app.example.com",
  "type": "A"
}
```

### Log Levels

| Level | Description |
|-------|-------------|
| `debug` | Detailed debugging information |
| `info` | Normal operational messages |
| `warn` | Warning conditions |
| `error` | Error conditions |

### Log Aggregation

For Loki:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: promtail-config
data:
  promtail.yaml: |
    scrape_configs:
      - job_name: external-dns-unifi
        kubernetes_sd_configs:
          - role: pod
        relabel_configs:
          - source_labels: [__meta_kubernetes_pod_label_app_kubernetes_io_name]
            action: keep
            regex: external-dns
        pipeline_stages:
          - json:
              expressions:
                level: level
                msg: msg
```
