# Environment Variables

All webhook configuration is done through environment variables.

## Required Variables

These variables must be set for the webhook to function.

### `WEBHOOK_UNIFI_HOST`

UniFi controller URL.

| | |
|---|---|
| **Required** | Yes |
| **Example** | `https://192.168.1.1` |

!!! warning "Use IP Address"
    Use the IP address of your UniFi controller, not a hostname like `unifi.local`. mDNS/Bonjour names may not resolve correctly in containerized environments.

### `WEBHOOK_UNIFI_API_KEY`

API key from your UniFi controller.

| | |
|---|---|
| **Required** | Yes |
| **Example** | `abc123...` |

Store this in a Kubernetes secret:

```yaml
env:
  - name: WEBHOOK_UNIFI_API_KEY
    valueFrom:
      secretKeyRef:
        name: unifi-credentials
        key: api-key
```

## Optional Variables

### UniFi Settings

#### `WEBHOOK_UNIFI_SITE`

UniFi site name. Most installations use "default".

| | |
|---|---|
| **Required** | No |
| **Default** | `default` |

#### `WEBHOOK_UNIFI_SKIP_TLS_VERIFY`

Skip TLS certificate verification for the UniFi controller connection.

| | |
|---|---|
| **Required** | No |
| **Default** | `true` |

!!! note
    Set to `true` for self-signed certificates (common with UniFi controllers). For production with valid certificates, set to `false`.

### Server Settings

#### `WEBHOOK_SERVER_HOST`

Bind address for the webhook API server.

| | |
|---|---|
| **Required** | No |
| **Default** | `localhost` |

The webhook API is called by external-dns running in the same pod, so `localhost` is appropriate.

#### `WEBHOOK_SERVER_PORT`

Port for the webhook API server.

| | |
|---|---|
| **Required** | No |
| **Default** | `8888` |

#### `WEBHOOK_HEALTH_HOST`

Bind address for the health/metrics server.

| | |
|---|---|
| **Required** | No |
| **Default** | `0.0.0.0` |

#### `WEBHOOK_HEALTH_PORT`

Port for the health/metrics server.

| | |
|---|---|
| **Required** | No |
| **Default** | `8080` |

### Logging Settings

#### `WEBHOOK_LOGGING_LEVEL`

Log verbosity level.

| | |
|---|---|
| **Required** | No |
| **Default** | `info` |
| **Values** | `debug`, `info`, `warn`, `error` |

#### `WEBHOOK_LOGGING_FORMAT`

Log output format.

| | |
|---|---|
| **Required** | No |
| **Default** | `json` |
| **Values** | `json`, `text` |

Use `json` for production (structured logs), `text` for development (human-readable).

### Debug Settings

#### `WEBHOOK_DEBUG_PPROF_ENABLED`

Enable pprof profiling endpoints.

| | |
|---|---|
| **Required** | No |
| **Default** | `false` |

!!! danger "Production Warning"
    **Never enable in production.** pprof endpoints expose sensitive runtime information and can impact performance.

#### `WEBHOOK_DEBUG_PPROF_PORT`

Port for pprof server when enabled.

| | |
|---|---|
| **Required** | No |
| **Default** | `6060` |

## Complete Example

```yaml
env:
  # Required
  - name: WEBHOOK_UNIFI_HOST
    value: "https://192.168.1.1"
  - name: WEBHOOK_UNIFI_API_KEY
    valueFrom:
      secretKeyRef:
        name: unifi-credentials
        key: api-key

  # Optional - UniFi
  - name: WEBHOOK_UNIFI_SITE
    value: "default"
  - name: WEBHOOK_UNIFI_SKIP_TLS_VERIFY
    value: "true"

  # Optional - Server
  - name: WEBHOOK_SERVER_HOST
    value: "localhost"
  - name: WEBHOOK_SERVER_PORT
    value: "8888"
  - name: WEBHOOK_HEALTH_HOST
    value: "0.0.0.0"
  - name: WEBHOOK_HEALTH_PORT
    value: "8080"

  # Optional - Logging
  - name: WEBHOOK_LOGGING_LEVEL
    value: "info"
  - name: WEBHOOK_LOGGING_FORMAT
    value: "json"
```
