# Configuration

This section covers all configuration options for external-dns-unifios-webhook.

<div class="grid cards" markdown>

-   :material-variable:{ .lg .middle } **Environment Variables**

    ---

    All available configuration options and their defaults.

    [:octicons-arrow-right-24: Environment Variables](environment.md)

-   :material-key:{ .lg .middle } **UniFi API Setup**

    ---

    Setting up API access to your UniFi controller.

    [:octicons-arrow-right-24: UniFi API Setup](unifi-api.md)

</div>

## Configuration Overview

The webhook is configured entirely through environment variables. No configuration files are required.

### Quick Reference

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `WEBHOOK_UNIFI_HOST` | Yes | - | UniFi controller URL |
| `WEBHOOK_UNIFI_API_KEY` | Yes | - | API key |
| `WEBHOOK_UNIFI_SITE` | No | `default` | Site name |
| `WEBHOOK_LOGGING_LEVEL` | No | `info` | Log level |

See [Environment Variables](environment.md) for the complete list.

## Kubernetes Secret Management

Store sensitive values in Kubernetes secrets:

```yaml
env:
  - name: WEBHOOK_UNIFI_API_KEY
    valueFrom:
      secretKeyRef:
        name: unifi-credentials
        key: api-key
```

!!! tip "External Secrets"
    Consider using [External Secrets Operator](https://external-secrets.io/) or [Sealed Secrets](https://sealed-secrets.netlify.app/) for production environments.
