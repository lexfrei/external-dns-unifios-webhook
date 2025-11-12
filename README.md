# external-dns-unifios-webhook

[![Go Version](https://img.shields.io/badge/go-1.25.4-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-BSD--3--Clause-blue.svg)](LICENSE)
[![Build Status](https://img.shields.io/github/actions/workflow/status/lexfrei/external-dns-unifios-webhook/ci.yml?branch=master)](https://github.com/lexfrei/external-dns-unifios-webhook/actions)

Webhook provider for [external-dns](https://github.com/kubernetes-sigs/external-dns) that integrates with UniFi OS DNS management. Enables automatic DNS record management in UniFi controllers (UniFi Dream Machine, Cloud Key, etc.) from Kubernetes.

## Features

- DNS record type support: A, AAAA, CNAME, TXT
- Domain filtering with include/exclude patterns and regex
- Kubernetes-ready health checks and readiness probes
- Prometheus metrics endpoint for monitoring
- Structured JSON logging with configurable levels
- OpenTelemetry instrumentation
- Lightweight container image built from scratch

## Requirements

### UniFi Controller

- UniFi OS Controller (UniFi Dream Machine, Cloud Key Gen2+, or UniFi Network Application)
- UniFi controller with API access enabled
- API key with DNS management permissions

### Creating UniFi API Key

1. Log in to UniFi controller web interface
2. Navigate to Settings → Admins
3. Select your admin user
4. Scroll to "API Access" section
5. Generate new API key
6. Save the key securely (shown only once)

### Optional

- Kubernetes cluster with external-dns deployed
- Prometheus for metrics collection

## Installation

### Kubernetes with Helm

Deploy external-dns with this webhook provider using the official external-dns Helm chart.

**Prerequisites:**
1. Add external-dns Helm repository:
   ```bash
   helm repo add external-dns https://kubernetes-sigs.github.io/external-dns/
   helm repo update
   ```

2. Create namespace and secret with UniFi API key:
   ```bash
   kubectl create namespace external-dns-unifi
   kubectl create secret generic unifi-credentials \
     --namespace external-dns-unifi \
     --from-literal=api-key=YOUR_UNIFI_API_KEY
   ```

3. Create `values.yaml` with webhook configuration:

This is the Helm values configuration for the external-dns chart. The `provider.webhook` section configures the webhook sidecar container that communicates with your UniFi controller:

```yaml
provider:
  name: webhook
  webhook:
    image:
      repository: ghcr.io/lexfrei/external-dns-unifios-webhook
      tag: latest
    env:
      - name: WEBHOOK_UNIFI_API_KEY
        valueFrom:
          secretKeyRef:
            name: unifi-credentials
            key: api-key
      - name: WEBHOOK_UNIFI_HOST
        value: "https://192.168.1.1"
    livenessProbe:
      httpGet:
        path: /healthz
        port: 8080
      initialDelaySeconds: 10
      periodSeconds: 10
    readinessProbe:
      httpGet:
        path: /readyz
        port: 8080
      initialDelaySeconds: 10
      periodSeconds: 10
```

1. Deploy external-dns with the webhook:

   ```bash
   helm upgrade --install external-dns-unifi external-dns/external-dns \
     --namespace external-dns-unifi \
     --values values.yaml
   ```

**Webhook configuration options:**

Required:

- `WEBHOOK_UNIFI_HOST`: UniFi controller URL (use IP address, not hostname like unifi.local)
- `WEBHOOK_UNIFI_API_KEY`: API key from UniFi controller (stored in Kubernetes secret)

Optional:

- `WEBHOOK_UNIFI_SITE`: UniFi site name (default: `default`)
- `WEBHOOK_UNIFI_SKIP_TLS_VERIFY`: Skip TLS certificate verification (default: `true`)
- `WEBHOOK_SERVER_HOST`: Webhook server bind address (default: `localhost`)
- `WEBHOOK_SERVER_PORT`: Webhook server port (default: `8888`)
- `WEBHOOK_HEALTH_HOST`: Health server bind address (default: `localhost`)
- `WEBHOOK_HEALTH_PORT`: Health server port (default: `8080`)
- `WEBHOOK_LOGGING_LEVEL`: Log level - `debug`, `info`, `warn`, `error` (default: `info`)
- `WEBHOOK_LOGGING_FORMAT`: Log format - `json` or `text` (default: `json`)

**External-DNS configuration options:**

- `domainFilters`: List of domains that external-dns will manage
- `policy`: Use `sync` to automatically create/delete records, or `upsert-only` to only create

For additional configuration options, see the [external-dns Helm chart documentation](https://github.com/kubernetes-sigs/external-dns/tree/master/charts/external-dns)

## Development

### Build

```bash
go build -o webhook ./cmd/webhook
```

### Run Locally

```bash
export WEBHOOK_UNIFI_HOST="192.168.1.1"
export WEBHOOK_UNIFI_API_KEY="your-api-key"
export WEBHOOK_LOGGING_LEVEL="debug"
go run ./cmd/webhook server
```

### Run Tests

```bash
go test ./...
```

### Build Container Image

```bash
podman build --tag external-dns-unifios-webhook:local --file Containerfile .
```

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

BSD-3-Clause License. See [LICENSE](LICENSE) file for details.

## Author

Aleksei Sviridkin <<f@lex.la>>
