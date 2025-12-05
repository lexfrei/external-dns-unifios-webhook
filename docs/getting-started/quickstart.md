# Quick Start

Get DNS records syncing from Kubernetes to your UniFi controller in 5 minutes.

## Prerequisites

Before starting, ensure you have:

- [x] UniFi controller with API key ([Prerequisites](prerequisites.md))
- [x] Webhook deployed ([Installation](installation.md))
- [x] A domain managed by your UniFi controller

## Create a Test Service

Deploy a simple service with DNS annotation:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: hello-world
  annotations:
    external-dns.alpha.kubernetes.io/hostname: hello.example.com
spec:
  type: LoadBalancer
  ports:
    - port: 80
      targetPort: 8080
  selector:
    app: hello-world
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello-world
spec:
  replicas: 1
  selector:
    matchLabels:
      app: hello-world
  template:
    metadata:
      labels:
        app: hello-world
    spec:
      containers:
        - name: hello
          image: nginx:alpine
          ports:
            - containerPort: 8080
```

Apply the manifest:

```bash
kubectl apply -f hello-world.yaml
```

## Verify DNS Record Creation

### Check external-dns Logs

```bash
kubectl logs -n external-dns-unifi -l app.kubernetes.io/name=external-dns
```

Look for entries indicating record creation:

```text
level=info msg="Creating record" record=hello.example.com type=A
```

### Check Webhook Logs

```bash
kubectl logs -n external-dns-unifi -l app.kubernetes.io/name=external-dns -c webhook
```

### Check UniFi Controller

1. Log in to your UniFi controller
2. Navigate to **Settings** → **Internet** → **DNS**
3. Verify the `hello.example.com` record appears

### Test DNS Resolution

```bash
dig hello.example.com @<unifi-controller-ip>
```

## Using Ingress

For Ingress resources, the hostname is automatically extracted:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: hello-ingress
spec:
  rules:
    - host: app.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: hello-world
                port:
                  number: 80
```

## Common Annotations

| Annotation | Description | Example |
|------------|-------------|---------|
| `external-dns.alpha.kubernetes.io/hostname` | DNS hostname | `app.example.com` |
| `external-dns.alpha.kubernetes.io/ttl` | Record TTL in seconds | `300` |
| `external-dns.alpha.kubernetes.io/target` | Override target IP/hostname | `10.0.0.1` |

## Multiple Hostnames

Specify multiple hostnames with comma separation:

```yaml
annotations:
  external-dns.alpha.kubernetes.io/hostname: app.example.com,www.example.com
```

## Cleanup

When you delete the Kubernetes resource, external-dns will automatically remove the DNS record from your UniFi controller (when using `policy: sync`).

```bash
kubectl delete -f hello-world.yaml
```

## Next Steps

- [Configuration](../configuration/index.md) - Customize webhook settings
- [Troubleshooting](../guides/troubleshooting.md) - Common issues and solutions
- [Monitoring](../guides/monitoring.md) - Set up metrics and alerting
