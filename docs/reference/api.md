# Webhook API

The webhook implements the external-dns webhook provider protocol.

## Overview

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/` | GET | Domain filter and negotiation |
| `/records` | GET | List DNS records |
| `/records` | POST | Apply DNS changes |
| `/adjustendpoints` | POST | Adjust endpoints |

## Base URLs

| Server | Port | Purpose |
|--------|------|---------|
| Webhook API | 8888 | DNS operations (localhost only) |
| Health | 8080 | Health checks and metrics |

## Endpoints

### GET /

Returns domain filter configuration and provider capabilities.

**Response:**

```json
{
  "domainFilter": {
    "filters": ["example.com"],
    "exclude": [],
    "regexFilter": "",
    "regexExclude": ""
  }
}
```

### GET /records

Returns all DNS records managed by the provider.

**Response:**

```json
{
  "endpoints": [
    {
      "dnsName": "app.example.com",
      "recordType": "A",
      "targets": ["10.0.0.1"],
      "recordTTL": 300
    }
  ]
}
```

### POST /records

Applies DNS record changes (create, update, delete).

**Request:**

```json
{
  "create": [
    {
      "dnsName": "new.example.com",
      "recordType": "A",
      "targets": ["10.0.0.2"],
      "recordTTL": 300
    }
  ],
  "updateOld": [],
  "updateNew": [],
  "delete": []
}
```

**Response:** `204 No Content`

### POST /adjustendpoints

Adjusts endpoints before external-dns processes them.

**Request:**

```json
{
  "endpoints": [
    {
      "dnsName": "app.example.com",
      "recordType": "A",
      "targets": ["10.0.0.1"]
    }
  ]
}
```

**Response:**

```json
{
  "endpoints": [
    {
      "dnsName": "app.example.com",
      "recordType": "A",
      "targets": ["10.0.0.1"]
    }
  ]
}
```

## Health Endpoints

### GET /healthz

Liveness probe. Returns 200 if process is alive.

**Response:**

```json
{"status": "ok"}
```

### GET /readyz

Readiness probe. Returns 200 if UniFi controller is reachable.

**Response:**

```json
{"status": "ready"}
```

### GET /metrics

Prometheus metrics endpoint.

**Response:** Prometheus text format

```
# HELP external_dns_unifi_records_managed Number of DNS records managed
# TYPE external_dns_unifi_records_managed gauge
external_dns_unifi_records_managed 42
```

## Error Responses

### 4xx Client Errors

```json
{
  "error": "invalid request",
  "message": "missing required field: dnsName"
}
```

!!! note
    external-dns does not retry 4xx errors (except 429).

### 5xx Server Errors

```json
{
  "error": "internal error",
  "message": "failed to connect to UniFi controller"
}
```

external-dns will retry 5xx errors.

## Request Limits

| Limit | Value |
|-------|-------|
| Request body size | 5 MB |
| Approximate record capacity | ~25,000 records |

## OpenAPI Specification

The API is defined using OpenAPI 3.0. Specifications are located in:

- `api/webhook/openapi.yaml` - Webhook API
- `api/health/openapi.yaml` - Health API

Generated code is in `api/*/generated.go`.
