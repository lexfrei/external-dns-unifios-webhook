# Troubleshooting

Common issues and solutions when using external-dns-unifios-webhook.

## Connection Issues

### Cannot Connect to UniFi Controller

**Symptoms:**

```
error="connection refused" host="https://192.168.1.1"
```

**Solutions:**

1. Verify the controller URL is correct
2. Check network connectivity from the pod:

    ```bash
    kubectl exec -it <pod> -c webhook -- wget -O- https://192.168.1.1
    ```

3. Ensure firewall allows HTTPS (443) traffic
4. Use IP address instead of hostname

### TLS Certificate Errors

**Symptoms:**

```
error="x509: certificate signed by unknown authority"
```

**Solutions:**

1. Enable TLS skip verification (default):

    ```yaml
    env:
      - name: WEBHOOK_UNIFI_SKIP_TLS_VERIFY
        value: "true"
    ```

2. Or mount the UniFi CA certificate

### Authentication Failed

**Symptoms:**

```
error="authentication failed" status=401
```

**Solutions:**

1. Verify API key is correct
2. Check the secret is properly mounted:

    ```bash
    kubectl get secret unifi-credentials -o jsonpath='{.data.api-key}' | base64 -d
    ```

3. Regenerate API key in UniFi controller
4. Ensure admin user is active

## DNS Record Issues

### Records Not Created

**Symptoms:**
DNS records don't appear in UniFi controller.

**Debugging:**

1. Check external-dns logs:

    ```bash
    kubectl logs -l app.kubernetes.io/name=external-dns
    ```

2. Check webhook logs:

    ```bash
    kubectl logs -l app.kubernetes.io/name=external-dns -c webhook
    ```

3. Verify domain filter matches:

    ```yaml
    domainFilters:
      - example.com  # Must match your hostname
    ```

### Wildcard CNAME Not Working

**Symptoms:**

```
error="wildcard CNAME not supported"
```

**Cause:**
UniFi uses dnsmasq which doesn't support wildcard CNAME records.

**Solution:**
Use individual A/AAAA records instead of wildcard CNAME.

### Duplicate CNAME Error

**Symptoms:**

```
error="duplicate CNAME record"
```

**Cause:**
dnsmasq doesn't support multiple CNAME records for the same name.

**Solution:**
Use A records with multiple targets for round-robin, or use a single CNAME.

## Performance Issues

### Slow DNS Updates

**Symptoms:**
Records take a long time to appear.

**Solutions:**

1. Reduce external-dns sync interval:

    ```yaml
    interval: 30s
    ```

2. Check for API rate limiting in webhook logs
3. Verify UniFi controller performance

### High Memory Usage

**Symptoms:**
Webhook container using excessive memory.

**Solutions:**

1. Check for large number of DNS records
2. Increase memory limits:

    ```yaml
    resources:
      limits:
        memory: 256Mi
    ```

3. Enable debug logging temporarily to investigate

## Health Check Failures

### Liveness Probe Failing

**Symptoms:**
Pod keeps restarting due to liveness probe failure.

**Debugging:**

```bash
kubectl describe pod <pod-name>
```

**Solutions:**

1. Increase `initialDelaySeconds`:

    ```yaml
    livenessProbe:
      initialDelaySeconds: 30
    ```

2. Check if UniFi controller is accessible

### Readiness Probe Failing

**Symptoms:**
Pod not receiving traffic.

**Debugging:**

```bash
kubectl port-forward <pod> 8080:8080
curl http://localhost:8080/readyz
```

**Solutions:**

1. Verify UniFi controller connectivity
2. Check webhook logs for errors

## Debugging

### Enable Debug Logging

```yaml
env:
  - name: WEBHOOK_LOGGING_LEVEL
    value: "debug"
```

### View All Logs

```bash
# external-dns logs
kubectl logs -l app.kubernetes.io/name=external-dns --tail=100

# webhook logs
kubectl logs -l app.kubernetes.io/name=external-dns -c webhook --tail=100
```

### Check Metrics

```bash
kubectl port-forward <pod> 8080:8080
curl http://localhost:8080/metrics | grep external_dns_unifi
```

## Getting Help

If you're still stuck:

1. Search [existing issues](https://github.com/lexfrei/external-dns-unifios-webhook/issues)
2. Enable debug logging and collect logs
3. [Open an issue](https://github.com/lexfrei/external-dns-unifios-webhook/issues/new) with:
    - Webhook version
    - UniFi OS and Network versions
    - Debug logs (redact API keys)
    - Steps to reproduce
