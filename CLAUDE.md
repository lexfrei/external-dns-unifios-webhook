# Development Guidelines for external-dns-unifios-webhook

## Current Development Mode

**IMPORTANT**: Currently developing WITHOUT git operations
- **NO commits** during active development
- **NO pushes** to remote
- Focus on iteration and testing
- Git operations only when explicitly requested by user

## Container Image Build and Deploy Workflow

### CRITICAL RULES
- **NEVER use `latest` tag in production**
- **ALWAYS use UUID-based tags for container images**
- **ALWAYS push image immediately after build**
- **ALWAYS update values.yaml with the new tag**

### Build and Deploy Process

1. **Generate UUID tag**:
```bash
UUID=$(uuidgen | tr '[:upper:]' '[:lower:]' | cut -d'-' -f1)
```

2. **Build image with UUID tag**:
```bash
podman build --platform linux/arm64 \
  --tag ghcr.io/lexfrei/external-dns-unifios-webhook:${UUID} \
  --file Containerfile .
```

3. **Push image immediately**:
```bash
podman push ghcr.io/lexfrei/external-dns-unifios-webhook:${UUID}
```

4. **Update values.yaml**:
```yaml
provider:
  webhook:
    image:
      repository: ghcr.io/lexfrei/external-dns-unifios-webhook
      tag: "${UUID}"  # Replace with actual UUID
```

5. **Apply with Helm**:
```bash
helm upgrade external-dns-unifi external-dns/external-dns \
  --namespace external-dns-unifi \
  --values deploy/kubernetes/external-dns-values.yaml
```

### Complete One-Liner for Quick Iteration

```bash
UUID=$(uuidgen | tr '[:upper:]' '[:lower:]' | cut -d'-' -f1) && \
podman build --platform linux/arm64 \
  --tag ghcr.io/lexfrei/external-dns-unifios-webhook:${UUID} \
  --file Containerfile . && \
podman push ghcr.io/lexfrei/external-dns-unifios-webhook:${UUID} && \
sed -i '' "s/tag: .*/tag: \"${UUID}\"/" deploy/kubernetes/external-dns-values.yaml && \
helm upgrade external-dns-unifi external-dns/external-dns \
  --namespace external-dns-unifi \
  --values deploy/kubernetes/external-dns-values.yaml && \
echo "Deployed with tag: ${UUID}"
```

## Why UUID Tags?

1. **Reproducibility**: Exact image version is always known
2. **No cache issues**: Kubernetes always pulls correct image
3. **Rollback capability**: Easy to revert to previous UUID tag
4. **Debugging**: Clear which code version is running
5. **Prevents confusion**: No wondering "is this the latest code?"

## GHCR Access

Images must be public or cluster must have imagePullSecrets configured.
Check image visibility: https://github.com/lexfrei/external-dns-unifios-webhook/pkgs/container/external-dns-unifios-webhook

## Testing Changes

After deploy, trigger reconciliation:
```bash
# Delete old DNS records to force recreation
kubectl delete pod -n external-dns-unifi -l app.kubernetes.io/name=external-dns

# Watch logs
kubectl logs -n external-dns-unifi -l app.kubernetes.io/name=external-dns \
  --container webhook --follow
```
