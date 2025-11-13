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

### Binary Release

Download the latest release from [GitHub Releases](https://github.com/lexfrei/external-dns-unifios-webhook/releases):

```bash
wget https://github.com/lexfrei/external-dns-unifios-webhook/releases/latest/download/external-dns-unifios-webhook-linux-amd64
chmod +x external-dns-unifios-webhook-linux-amd64
mv external-dns-unifios-webhook-linux-amd64 /usr/local/bin/external-dns-unifios-webhook
```

### Container Image

Pull from GitHub Container Registry:

```bash
docker pull ghcr.io/lexfrei/external-dns-unifios-webhook:latest
```

Or using Podman:

```bash
podman pull ghcr.io/lexfrei/external-dns-unifios-webhook:latest
```

### Build from Source

```bash
git clone https://github.com/lexfrei/external-dns-unifios-webhook.git
cd external-dns-unifios-webhook
go build -o webhook ./cmd/webhook
```

## Configuration

Configuration can be provided via environment variables, CLI flags, or YAML config file.

### Environment Variables

All environment variables use `WEBHOOK_` prefix:

| Variable | Description | Default |
|----------|-------------|---------|
| `WEBHOOK_UNIFI_HOST` | UniFi controller hostname or IP | - |
| `WEBHOOK_UNIFI_API_KEY` | UniFi API key | - |
| `WEBHOOK_UNIFI_SITE` | UniFi site name | `default` |
| `WEBHOOK_SERVER_HOST` | Webhook server bind address | `localhost` |
| `WEBHOOK_SERVER_PORT` | Webhook server port | `8888` |
| `WEBHOOK_HEALTH_HOST` | Health/metrics server bind address | `0.0.0.0` |
| `WEBHOOK_HEALTH_PORT` | Health/metrics server port | `8080` |
| `WEBHOOK_LOGGING_LEVEL` | Log level (debug, info, warn, error) | `info` |
| `WEBHOOK_LOGGING_FORMAT` | Log format (json, text) | `json` |
| `WEBHOOK_DOMAIN_FILTER_FILTERS` | Comma-separated domain filters | - |
| `WEBHOOK_DOMAIN_FILTER_EXCLUDE_FILTERS` | Comma-separated exclude filters | - |
| `WEBHOOK_DOMAIN_FILTER_REGEX_FILTERS` | Comma-separated regex filters | - |
| `WEBHOOK_DOMAIN_FILTER_REGEX_EXCLUDE_FILTERS` | Comma-separated regex exclude filters | - |

### CLI Flags

```bash
external-dns-unifios-webhook server --help
```

### Config File

Create `config.yaml` in one of these locations:
- Current directory: `./config.yaml`
- System config: `/etc/external-dns-unifios-webhook/config.yaml`
- User home: `$HOME/.external-dns-unifios-webhook/config.yaml`

Example configuration:

```yaml
unifi:
  host: "192.168.1.1"
  api_key: "your-api-key-here"
  site: "default"

server:
  host: "localhost"
  port: "8888"

health:
  host: "0.0.0.0"
  port: "8080"

domain_filter:
  filters:
    - "example.com"
    - "internal.lan"
  exclude_filters:
    - "skip.example.com"

logging:
  level: "info"
  format: "json"
```

## Usage

### Standalone

```bash
export WEBHOOK_UNIFI_HOST="192.168.1.1"
export WEBHOOK_UNIFI_API_KEY="your-api-key"
external-dns-unifios-webhook server
```

### Docker

```bash
docker run --rm \
  --publish 8888:8888 \
  --publish 8080:8080 \
  --env WEBHOOK_UNIFI_HOST="192.168.1.1" \
  --env WEBHOOK_UNIFI_API_KEY="your-api-key" \
  ghcr.io/lexfrei/external-dns-unifios-webhook:latest \
  server
```

### Podman

```bash
podman run --rm \
  --publish 8888:8888 \
  --publish 8080:8080 \
  --env WEBHOOK_UNIFI_HOST="192.168.1.1" \
  --env WEBHOOK_UNIFI_API_KEY="your-api-key" \
  ghcr.io/lexfrei/external-dns-unifios-webhook:latest \
  server
```

### Kubernetes Deployment

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: unifios-webhook-secret
  namespace: external-dns
type: Opaque
stringData:
  api-key: "your-unifi-api-key"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: external-dns-unifios-webhook
  namespace: external-dns
spec:
  replicas: 1
  selector:
    matchLabels:
      app: external-dns-unifios-webhook
  template:
    metadata:
      labels:
        app: external-dns-unifios-webhook
    spec:
      containers:
      - name: webhook
        image: ghcr.io/lexfrei/external-dns-unifios-webhook:latest
        args:
        - server
        env:
        - name: WEBHOOK_UNIFI_HOST
          value: "192.168.1.1"
        - name: WEBHOOK_UNIFI_API_KEY
          valueFrom:
            secretKeyRef:
              name: unifios-webhook-secret
              key: api-key
        - name: WEBHOOK_UNIFI_SITE
          value: "default"
        - name: WEBHOOK_SERVER_HOST
          value: "0.0.0.0"
        - name: WEBHOOK_DOMAIN_FILTER_FILTERS
          value: "example.com,internal.lan"
        ports:
        - containerPort: 8888
          name: webhook
          protocol: TCP
        - containerPort: 8080
          name: health
          protocol: TCP
        livenessProbe:
          httpGet:
            path: /health
            port: health
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /ready
            port: health
          initialDelaySeconds: 5
          periodSeconds: 10
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "128Mi"
            cpu: "200m"
---
apiVersion: v1
kind: Service
metadata:
  name: external-dns-unifios-webhook
  namespace: external-dns
spec:
  selector:
    app: external-dns-unifios-webhook
  ports:
  - name: webhook
    port: 8888
    targetPort: webhook
  - name: health
    port: 8080
    targetPort: health
```

### External-DNS Configuration

Configure external-dns to use this webhook provider:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: external-dns
  namespace: external-dns
spec:
  replicas: 1
  selector:
    matchLabels:
      app: external-dns
  template:
    metadata:
      labels:
        app: external-dns
    spec:
      serviceAccountName: external-dns
      containers:
      - name: external-dns
        image: registry.k8s.io/external-dns/external-dns:v0.15.0
        args:
        - --source=service
        - --source=ingress
        - --provider=webhook
        - --webhook-provider-url=http://external-dns-unifios-webhook.external-dns.svc.cluster.local:8888
        - --domain-filter=example.com
        - --log-level=info
```

## API Endpoints

### Webhook API (port 8888)

The webhook implements the external-dns webhook provider specification:

- `GET /` - Returns domain filter configuration
- `GET /records` - Lists all managed DNS records
- `POST /records` - Applies DNS record changes
- `POST /adjustendpoints` - Adjusts endpoints (pass-through)

### Health and Metrics API (port 8080)

- `GET /health` - Liveness probe endpoint
- `GET /ready` - Readiness probe endpoint
- `GET /metrics` - Prometheus metrics

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
