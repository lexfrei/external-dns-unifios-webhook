# UniFi API Setup

This guide covers setting up API access to your UniFi controller for the webhook.

## Creating an API Key

### Step 1: Access Admin Settings

1. Log in to your UniFi controller web interface
2. Navigate to **Settings** (gear icon)
3. Select **Admins** from the sidebar

### Step 2: Select Admin User

Select the admin user that will be used for API access.

!!! tip "Dedicated User"
    For production, create a dedicated admin user for the webhook rather than using your personal account.

### Step 3: Generate API Key

1. Scroll to the **API Access** section
2. Click **Generate New API Key**
3. Give the key a descriptive name (e.g., "external-dns-webhook")
4. Copy the key immediately - it's shown only once

!!! warning "Save Immediately"
    The API key is displayed only once. If you lose it, you'll need to generate a new one.

## API Key Permissions

The API key inherits all permissions from the admin user account. There is no granular permission control for API keys in UniFi.

!!! note "Security Consideration"
    Consider creating a dedicated admin user for the webhook to isolate access and simplify key rotation.

## Storing the API Key

### Kubernetes Secret

```bash
kubectl create secret generic unifi-credentials \
  --namespace external-dns-unifi \
  --from-literal=api-key=YOUR_API_KEY
```

Reference in deployment:

```yaml
env:
  - name: WEBHOOK_UNIFI_API_KEY
    valueFrom:
      secretKeyRef:
        name: unifi-credentials
        key: api-key
```

### External Secrets

Using External Secrets Operator with AWS Secrets Manager:

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: unifi-credentials
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secrets
    kind: ClusterSecretStore
  target:
    name: unifi-credentials
  data:
    - secretKey: api-key
      remoteRef:
        key: unifi/external-dns
        property: api-key
```

## TLS Configuration

UniFi controllers typically use self-signed certificates.

### Skip Verification (Default)

By default, TLS verification is skipped:

```yaml
env:
  - name: WEBHOOK_UNIFI_SKIP_TLS_VERIFY
    value: "true"
```

### Custom CA (Advanced)

To use proper TLS verification:

1. Export your UniFi controller's CA certificate
2. Mount it in the container
3. Set `WEBHOOK_UNIFI_SKIP_TLS_VERIFY=false`
4. Configure the CA in the container's trust store

## Multi-Site Configuration

If your UniFi controller manages multiple sites, specify the target site:

```yaml
env:
  - name: WEBHOOK_UNIFI_SITE
    value: "my-site-name"
```

To find your site name:

1. Log in to UniFi controller
2. Navigate to the site
3. Check the URL: `https://controller/network/site/SITE_NAME/...`

## Troubleshooting

### Authentication Failed

```text
error="authentication failed" status=401
```

- Verify API key is correct
- Check API key hasn't expired
- Ensure admin user is active

### Connection Refused

```text
error="connection refused"
```

- Verify controller URL is accessible
- Check firewall rules
- Use IP address instead of hostname

### Permission Denied

```text
error="permission denied" status=403
```

- Verify admin user is active
- Check site access (if using multi-site)
- Confirm API key belongs to correct user
