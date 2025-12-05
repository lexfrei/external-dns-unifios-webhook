# Development

Resources for contributing to external-dns-unifios-webhook.

<div class="grid cards" markdown>

-   :material-laptop:{ .lg .middle } **Development Setup**

    ---

    Set up your local development environment.

    [:octicons-arrow-right-24: Development Setup](setup.md)

-   :material-sitemap:{ .lg .middle } **Architecture**

    ---

    Understand the system design.

    [:octicons-arrow-right-24: Architecture](architecture.md)

-   :material-source-pull:{ .lg .middle } **Contributing**

    ---

    Guidelines for contributing.

    [:octicons-arrow-right-24: Contributing](contributing.md)

</div>

## Quick Start

```bash
# Clone
git clone https://github.com/lexfrei/external-dns-unifios-webhook.git
cd external-dns-unifios-webhook

# Build
go build -o webhook ./cmd/webhook

# Test
go test ./...

# Run locally
export WEBHOOK_UNIFI_HOST="https://192.168.1.1"
export WEBHOOK_UNIFI_API_KEY="your-api-key"
./webhook
```

## Requirements

- Go 1.25+
- Access to a UniFi controller (for integration testing)
- Docker/Podman (for container builds)
