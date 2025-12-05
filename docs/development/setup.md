# Development Setup

Set up your local environment for developing external-dns-unifios-webhook.

## Prerequisites

- Go 1.25 or later
- Git
- Access to a UniFi controller (for integration testing)
- Podman or Docker (for container builds)

## Clone Repository

```bash
git clone https://github.com/lexfrei/external-dns-unifios-webhook.git
cd external-dns-unifios-webhook
```

## Install Dependencies

```bash
go mod download
```

## Build

### Binary

```bash
go build -o webhook ./cmd/webhook
```

### With Version Info

```bash
go build -ldflags "-s -w -X main.Version=dev -X main.Gitsha=$(git rev-parse HEAD)" \
  -trimpath -o webhook ./cmd/webhook
```

## Run Locally

Set required environment variables:

```bash
export WEBHOOK_UNIFI_HOST="https://192.168.1.1"
export WEBHOOK_UNIFI_API_KEY="your-api-key"
export WEBHOOK_LOGGING_LEVEL="debug"
```

Run the webhook:

```bash
./webhook
```

The webhook will listen on:

- `localhost:8888` - Webhook API
- `0.0.0.0:8080` - Health/metrics endpoints

## Testing

### Run All Tests

```bash
go test ./...
```

### With Race Detection

```bash
go test -race ./...
```

### With Coverage

```bash
go test -race -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -html=coverage.out
```

### Specific Package

```bash
go test ./internal/provider/...
```

## Linting

### Install golangci-lint

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Run Linter

```bash
golangci-lint run
```

### Auto-fix

```bash
golangci-lint run --fix
```

## Container Build

### Build Image

```bash
podman build --tag external-dns-unifios-webhook:local --file Containerfile .
```

### Run Container

```bash
podman run --rm \
  --env WEBHOOK_UNIFI_HOST="https://192.168.1.1" \
  --env WEBHOOK_UNIFI_API_KEY="your-api-key" \
  --publish 8888:8888 \
  --publish 8080:8080 \
  external-dns-unifios-webhook:local
```

## IDE Setup

### VS Code

Recommended extensions:

- Go (golang.go)
- EditorConfig for VS Code
- YAML

Settings (`.vscode/settings.json`):

```json
{
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"],
  "go.testFlags": ["-race"],
  "editor.formatOnSave": true
}
```

### GoLand

1. Enable Go modules integration
2. Set Go to use golangci-lint as external linter
3. Enable "Format on save"

## Environment Variables for Development

| Variable | Description | Example |
|----------|-------------|---------|
| `WEBHOOK_UNIFI_HOST` | UniFi controller URL | `https://192.168.1.1` |
| `WEBHOOK_UNIFI_API_KEY` | API key | `abc123...` |
| `WEBHOOK_LOGGING_LEVEL` | Log level | `debug` |
| `WEBHOOK_LOGGING_FORMAT` | Log format | `text` |

## Debugging

### Enable pprof

```bash
export WEBHOOK_DEBUG_PPROF_ENABLED=true
./webhook
```

Access pprof at `http://localhost:6060/debug/pprof/`

### Debug Logging

```bash
export WEBHOOK_LOGGING_LEVEL=debug
export WEBHOOK_LOGGING_FORMAT=text
./webhook
```
