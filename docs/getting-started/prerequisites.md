# Prerequisites

Before installing external-dns-unifios-webhook, ensure your environment meets the following requirements.

## Version Requirements

| Component | Minimum Version | Notes |
|-----------|-----------------|-------|
| external-dns | v0.20.0 | Webhook provider support |
| UniFi OS | 4.3.9 | API compatibility |
| UniFi Network | 9.4.19 | DNS management features |
| Kubernetes | 1.25+ | Recommended |
| Helm | 3.x | For Helm installation |

## UniFi Controller Requirements

### Supported Controllers

- **UniFi Dream Machine** (UDM, UDM Pro, UDM SE)
- **UniFi Cloud Key Gen2+**
- **UniFi Network Application** (self-hosted)

!!! warning "Older Controllers"
    UniFi Cloud Key Gen1 and older controllers running UniFi Network Application < 9.4.19 are not supported due to API differences.

### Network Access

The webhook needs network access to your UniFi controller:

- HTTPS access to controller (default port 443)
- Controller must be reachable from the Kubernetes cluster
- Use IP address instead of hostname (e.g., `https://192.168.1.1` not `https://unifi.local`)

## Creating UniFi API Key

The webhook requires an API key to authenticate with your UniFi controller.

### Steps

1. Log in to your UniFi controller web interface
2. Navigate to **Settings** → **Admins**
3. Select your admin user
4. Scroll to the **API Access** section
5. Click **Generate New API Key**
6. Copy and save the key securely (it's shown only once)

!!! important "API Key Permissions"
    The API key inherits permissions from the admin user account. Ensure the user has sufficient privileges to manage DNS records.

### Recommended Setup

For production environments, create a dedicated admin user for the webhook:

1. Create a new admin user (e.g., `external-dns-webhook`)
2. Assign minimal required permissions (DNS management)
3. Generate an API key for this user
4. Store the key securely in Kubernetes secrets

## Kubernetes Requirements

### external-dns

external-dns must be deployed with webhook provider support:

```yaml
provider:
  name: webhook
```

If you're using an older external-dns configuration with provider sources, you'll need to migrate to the webhook provider configuration.

### Resources

Recommended resource allocations:

```yaml
resources:
  requests:
    cpu: 10m
    memory: 32Mi
  limits:
    cpu: 100m
    memory: 128Mi
```

## Optional Components

### Prometheus

For metrics collection, ensure Prometheus can scrape the webhook's metrics endpoint:

- Port: `8080` (default)
- Path: `/metrics`

### Network Policies

If using network policies, allow:

- Egress from webhook to UniFi controller (HTTPS/443)
- Ingress to webhook from external-dns (HTTP/8888)
- Ingress to health endpoint (HTTP/8080)

## Next Steps

Once prerequisites are met, proceed to [Installation](installation.md).
