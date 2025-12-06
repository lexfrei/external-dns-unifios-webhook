# DNS Record Types

Supported DNS record types and their behavior in external-dns-unifios-webhook.

## Supported Types

| Type | Supported | Notes |
|------|-----------|-------|
| A | Yes | IPv4 address records |
| AAAA | Yes | IPv6 address records |
| CNAME | Yes | Canonical name records |
| TXT | Yes | Text records (no TTL support) |

## Record Details

### A Records

IPv4 address records.

```yaml
annotations:
  external-dns.alpha.kubernetes.io/hostname: app.example.com
  external-dns.alpha.kubernetes.io/target: 10.0.0.1
  external-dns.alpha.kubernetes.io/ttl: "300"
```

**Multi-target (Round Robin):**

```yaml
annotations:
  external-dns.alpha.kubernetes.io/hostname: app.example.com
  external-dns.alpha.kubernetes.io/target: "10.0.0.1,10.0.0.2"
```

Creates separate DNS records for each target, enabling round-robin DNS.

### AAAA Records

IPv6 address records.

```yaml
annotations:
  external-dns.alpha.kubernetes.io/hostname: app.example.com
  external-dns.alpha.kubernetes.io/target: "2001:db8::1"
```

### CNAME Records

Canonical name records pointing to another hostname.

```yaml
annotations:
  external-dns.alpha.kubernetes.io/hostname: www.example.com
  external-dns.alpha.kubernetes.io/target: app.example.com
```

!!! warning "Limitations"
    - Wildcard CNAME (`*.example.com`) not supported
    - Duplicate CNAME records for same name not supported
    - These are dnsmasq limitations

### TXT Records

Text records for verification and metadata.

```yaml
annotations:
  external-dns.alpha.kubernetes.io/hostname: _verify.example.com
  external-dns.alpha.kubernetes.io/target: "verification-token"
```

!!! note "No TTL Support"
    UniFi API does not support TTL for TXT records. TTL annotations are ignored.

## TTL Behavior

### Default TTL

Default TTL is 300 seconds (5 minutes).

### Custom TTL

Set TTL via annotation:

```yaml
annotations:
  external-dns.alpha.kubernetes.io/ttl: "600"
```

### TTL Limitations

| Record Type | TTL Support |
|-------------|-------------|
| A | Yes |
| AAAA | Yes |
| CNAME | Yes |
| TXT | No |

## Record Ownership

external-dns uses TXT records to track ownership:

```text
_externaldns.app.example.com TXT "heritage=external-dns,external-dns/owner=..."
```

This prevents conflicts when multiple external-dns instances manage the same domain.

### Owner ID

Configure unique owner per cluster:

```yaml
txtOwnerId: cluster-production
```

### TXT Prefix

Configure TXT record prefix:

```yaml
txtPrefix: "_externaldns."
```

## UniFi-Specific Behavior

### Record Creation

Each target creates a separate DNS record in UniFi:

```yaml
# This annotation:
external-dns.alpha.kubernetes.io/target: "10.0.0.1,10.0.0.2"

# Creates two records in UniFi:
# app.example.com A 10.0.0.1
# app.example.com A 10.0.0.2
```

### Record Updates

Updates are performed as delete + create operations.

### Record Deletion

Records are deleted when:

- Kubernetes resource is deleted (with `policy: sync`)
- Annotation is removed
- Service loses its external IP

## Best Practices

1. **Use unique owner IDs** per cluster to prevent conflicts
2. **Avoid wildcard CNAME** - use individual records
3. **Set appropriate TTL** for your use case
4. **Use domain filters** to limit scope
5. **Test in staging** before production deployment
