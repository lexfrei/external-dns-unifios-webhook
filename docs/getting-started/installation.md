# Installation

This guide covers deploying external-dns-unifios-webhook to your Kubernetes cluster.

## Helm Installation (Recommended)

The webhook is designed to work with the official external-dns Helm chart as a sidecar container.

### Add Helm Repository

```bash
helm repo add external-dns https://kubernetes-sigs.github.io/external-dns/
helm repo update
```

### Create Namespace and Secret

```bash
kubectl create namespace external-dns-unifi

kubectl create secret generic unifi-credentials \
  --namespace external-dns-unifi \
  --from-literal=api-key=YOUR_UNIFI_API_KEY
```

!!! warning "Security"
    Never commit API keys to version control. Use Kubernetes secrets, sealed secrets, or external secret management.

### Create Values File

Create `values.yaml` with webhook configuration:

```yaml
# Webhook provider configuration
provider:
  name: webhook
  webhook:
    image:
      repository: ghcr.io/lexfrei/external-dns-unifios-webhook
      tag: latest  # Pin to specific version in production
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
    resources:
      requests:
        cpu: 10m
        memory: 32Mi
      limits:
        cpu: 100m
        memory: 128Mi

# Configure external-dns behavior
# Adjust these according to your needs
domainFilters:
  - example.com
policy: sync
interval: 1m
```

### Deploy

```bash
helm upgrade --install external-dns-unifi external-dns/external-dns \
  --namespace external-dns-unifi \
  --values values.yaml
```

### Verify Deployment

```bash
# Check pods are running
kubectl get pods -n external-dns-unifi

# Check logs
kubectl logs -n external-dns-unifi -l app.kubernetes.io/name=external-dns -c webhook
```

## Manual Installation

For environments without Helm, you can deploy manually.

### Create Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: external-dns-unifi
  namespace: external-dns-unifi
spec:
  replicas: 1
  selector:
    matchLabels:
      app: external-dns-unifi
  template:
    metadata:
      labels:
        app: external-dns-unifi
    spec:
      containers:
        - name: external-dns
          image: registry.k8s.io/external-dns/external-dns:v0.20.0
          args:
            - --source=service
            - --source=ingress
            - --provider=webhook
            - --webhook-provider-url=http://localhost:8888
          ports:
            - containerPort: 7979
              name: http
        - name: webhook
          image: ghcr.io/lexfrei/external-dns-unifios-webhook:latest
          env:
            - name: WEBHOOK_UNIFI_HOST
              value: "https://192.168.1.1"
            - name: WEBHOOK_UNIFI_API_KEY
              valueFrom:
                secretKeyRef:
                  name: unifi-credentials
                  key: api-key
          ports:
            - containerPort: 8888
              name: webhook
            - containerPort: 8080
              name: health
          livenessProbe:
            httpGet:
              path: /healthz
              port: 8080
            initialDelaySeconds: 10
          readinessProbe:
            httpGet:
              path: /readyz
              port: 8080
            initialDelaySeconds: 10
```

## Container Image

The webhook is available as a multi-arch container image:

```
ghcr.io/lexfrei/external-dns-unifios-webhook:latest
```

### Supported Architectures

- `linux/amd64`
- `linux/arm64`

### Image Tags

| Tag | Description |
|-----|-------------|
| `latest` | Latest stable release |
| `vX.Y.Z` | Specific version |
| `X.Y` | Latest patch for major.minor |
| `X` | Latest minor for major |

!!! tip "Production"
    Always pin to a specific version tag in production to avoid unexpected updates.

## Verification

After installation, verify the webhook is working:

1. Check pod status:

    ```bash
    kubectl get pods -n external-dns-unifi
    ```

2. Check webhook logs:

    ```bash
    kubectl logs -n external-dns-unifi -l app.kubernetes.io/name=external-dns -c webhook
    ```

3. Check health endpoint:

    ```bash
    kubectl port-forward -n external-dns-unifi svc/external-dns-unifi 8080:8080
    curl http://localhost:8080/healthz
    ```

## Next Steps

- [Quick Start](quickstart.md) - Create your first DNS record
- [Configuration](../configuration/index.md) - Customize webhook behavior
