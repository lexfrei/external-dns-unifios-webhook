# Security

Security considerations and best practices for external-dns-unifios-webhook.

## Authentication

### API Key Security

The UniFi API key grants access to DNS management. Protect it carefully.

**Best Practices:**

1. Store in Kubernetes secrets, not ConfigMaps
2. Use RBAC to limit secret access
3. Consider external secret management (Vault, AWS Secrets Manager)
4. Rotate keys periodically
5. Use dedicated admin user with minimal permissions

**Example:**

```yaml
env:
  - name: WEBHOOK_UNIFI_API_KEY
    valueFrom:
      secretKeyRef:
        name: unifi-credentials
        key: api-key
```

**Do NOT:**

```yaml
# Never do this!
env:
  - name: WEBHOOK_UNIFI_API_KEY
    value: "your-api-key-in-plain-text"
```

## Network Security

### TLS

The webhook connects to UniFi controller over HTTPS.

**Self-signed certificates:**

By default, TLS verification is skipped for self-signed certificates:

```yaml
env:
  - name: WEBHOOK_UNIFI_SKIP_TLS_VERIFY
    value: "true"
```

**Production with valid certificates:**

```yaml
env:
  - name: WEBHOOK_UNIFI_SKIP_TLS_VERIFY
    value: "false"
```

### Network Policies

Restrict network access:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: external-dns-unifi
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/name: external-dns
  policyTypes:
    - Egress
    - Ingress
  ingress:
    # Allow health checks from kubelet
    - ports:
        - port: 8080
  egress:
    # Allow DNS
    - to:
        - namespaceSelector: {}
      ports:
        - port: 53
          protocol: UDP
    # Allow UniFi controller
    - to:
        - ipBlock:
            cidr: 192.168.1.1/32
      ports:
        - port: 443
```

## Container Security

### Non-root User

The container runs as non-root user (65534/nobody):

```dockerfile
USER 65534:65534
```

### Read-only Filesystem

The container can run with read-only root filesystem:

```yaml
securityContext:
  readOnlyRootFilesystem: true
```

### Minimal Image

Built from `scratch` with only:

- The webhook binary
- CA certificates

No shell, no package manager, minimal attack surface.

### Security Context

Recommended pod security context:

```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 65534
  runAsGroup: 65534
  fsGroup: 65534
  seccompProfile:
    type: RuntimeDefault

containers:
  - name: webhook
    securityContext:
      allowPrivilegeEscalation: false
      readOnlyRootFilesystem: true
      capabilities:
        drop:
          - ALL
```

## Secrets Management

### Kubernetes Secrets

Basic approach - use Kubernetes secrets:

```bash
kubectl create secret generic unifi-credentials \
  --from-literal=api-key=YOUR_API_KEY
```

### Sealed Secrets

For GitOps with encrypted secrets:

```yaml
apiVersion: bitnami.com/v1alpha1
kind: SealedSecret
metadata:
  name: unifi-credentials
spec:
  encryptedData:
    api-key: AgBy3i...
```

### External Secrets Operator

For centralized secret management:

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: unifi-credentials
spec:
  secretStoreRef:
    name: vault
    kind: ClusterSecretStore
  target:
    name: unifi-credentials
  data:
    - secretKey: api-key
      remoteRef:
        key: secret/data/unifi
        property: api-key
```

## RBAC

### Minimal Permissions

The webhook only needs:

- Read/write access to DNS records via UniFi API
- No Kubernetes API access required (external-dns handles this)

### Service Account

If using Kubernetes secrets:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: external-dns-unifi
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: external-dns-unifi
rules:
  - apiGroups: [""]
    resources: ["secrets"]
    resourceNames: ["unifi-credentials"]
    verbs: ["get"]
```

## Audit Logging

### Webhook Logs

Enable info or debug logging to audit DNS operations:

```yaml
env:
  - name: WEBHOOK_LOGGING_LEVEL
    value: "info"
```

Log output includes:

- Record creation/deletion events
- Error conditions
- Connection status

### UniFi Controller Logs

The UniFi controller also logs API operations.

## Vulnerability Scanning

### Container Images

Images are scanned with Trivy in CI:

```yaml
- name: Run Trivy vulnerability scanner
  uses: aquasecurity/trivy-action@master
  with:
    scan-type: 'fs'
    severity: 'CRITICAL,HIGH,MEDIUM'
```

### Dependencies

Dependencies are monitored with Dependabot/Renovate.

## Reporting Security Issues

Report security vulnerabilities via:

- Email: security@lex.la
- GitHub Security Advisories

Do NOT report security issues in public GitHub issues.

See [SECURITY.md](https://github.com/lexfrei/external-dns-unifios-webhook/blob/master/SECURITY.md) for full policy.
