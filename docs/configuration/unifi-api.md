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

The API key inherits permissions from the admin user account. The webhook requires:

- **Read** access to DNS records
- **Write** access to DNS records
- Access to the configured site

### Minimal Permissions

For security, create an admin user with minimal required permissions:

1. Create a new admin user
2. Set role to **Limited Admin**
3. Enable only DNS-related permissions
4. Generate API key for this user

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

```
error="authentication failed" status=401
```

- Verify API key is correct
- Check API key hasn't expired
- Ensure admin user is active

### Connection Refused

```
error="connection refused"
```

- Verify controller URL is accessible
- Check firewall rules
- Use IP address instead of hostname

### Permission Denied

```
error="permission denied" status=403
```

- Verify admin user has DNS permissions
- Check site access permissions
- Confirm API key belongs to correct user
