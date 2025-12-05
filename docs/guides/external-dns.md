# external-dns Integration

This guide covers configuring external-dns to work with the UniFi webhook.

## Webhook Provider Configuration

external-dns v0.20.0+ supports webhook providers. Configure external-dns to use the webhook:

```yaml
provider:
  name: webhook
  webhook:
    image:
      repository: ghcr.io/lexfrei/external-dns-unifios-webhook
      tag: latest
```

## Source Configuration

### Services

```yaml
sources:
  - service
```

Annotate services:

```yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    external-dns.alpha.kubernetes.io/hostname: app.example.com
spec:
  type: LoadBalancer
```

### Ingress

```yaml
sources:
  - ingress
```

Hostnames are extracted automatically from Ingress rules:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: app
spec:
  rules:
    - host: app.example.com  # Automatically registered
```

### CRDs

For Gateway API or custom resources:

```yaml
sources:
  - crd
```

## Domain Filtering

### Include Domains

Only manage specific domains:

```yaml
domainFilters:
  - example.com
  - internal.local
```

### Exclude Domains

Exclude specific domains:

```yaml
excludeDomains:
  - staging.example.com
```

### Regex Filtering

Use regex for complex patterns:

```yaml
regexDomainFilter: ".*\\.prod\\.example\\.com$"
regexDomainExclusion: ".*\\.test\\.example\\.com$"
```

## Sync Policies

### Sync (Recommended)

Creates, updates, and deletes records:

```yaml
policy: sync
```

### Upsert-Only

Only creates and updates, never deletes:

```yaml
policy: upsert-only
```

## TTL Configuration

### Default TTL

Set default TTL for all records:

```yaml
txtOwnerId: external-dns-unifi
txtPrefix: "_externaldns."
```

### Per-Record TTL

Override TTL with annotation:

```yaml
annotations:
  external-dns.alpha.kubernetes.io/ttl: "300"
```

!!! note "TXT Record Limitation"
    UniFi does not support TTL for TXT records. TTL annotations are ignored for TXT records.

## Interval Configuration

How often external-dns syncs:

```yaml
interval: 1m  # Default
```

For busy clusters, consider longer intervals:

```yaml
interval: 5m
```

## Complete Example

```yaml
provider:
  name: webhook
  webhook:
    image:
      repository: ghcr.io/lexfrei/external-dns-unifios-webhook
      tag: v1.0.0
    env:
      - name: WEBHOOK_UNIFI_HOST
        value: "https://192.168.1.1"
      - name: WEBHOOK_UNIFI_API_KEY
        valueFrom:
          secretKeyRef:
            name: unifi-credentials
            key: api-key

sources:
  - service
  - ingress

domainFilters:
  - example.com
  - home.local

policy: sync
interval: 1m

txtOwnerId: external-dns-unifi
txtPrefix: "_externaldns."
```

## Multiple Clusters

When running external-dns in multiple clusters pointing to the same UniFi controller:

1. Use unique `txtOwnerId` per cluster:

    ```yaml
    # Cluster A
    txtOwnerId: cluster-a

    # Cluster B
    txtOwnerId: cluster-b
    ```

2. Consider domain partitioning:

    ```yaml
    # Cluster A
    domainFilters:
      - a.example.com

    # Cluster B
    domainFilters:
      - b.example.com
    ```
