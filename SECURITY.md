# Security Policy

## Supported Versions

We release patches for security vulnerabilities for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |
| < 0.1   | :x:                |

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

If you discover a security vulnerability, please send an email to:

**f@lex.la**

Please include the following information in your report:

- Description of the vulnerability
- Steps to reproduce the issue
- Potential impact
- Suggested fix (if available)

### What to Expect

- **Acknowledgment**: You will receive a response within 48 hours acknowledging your report
- **Assessment**: We will investigate and assess the vulnerability
- **Updates**: You will be kept informed of the progress
- **Disclosure**: Once fixed, we will coordinate disclosure timing with you
- **Credit**: Security researchers will be credited (unless you prefer to remain anonymous)

## Security Best Practices

When deploying this webhook:

1. **Use HTTPS**: Always use HTTPS for UniFi controller connections
2. **Secure API Keys**: Store API keys in Kubernetes secrets, never in configuration files
3. **Network Policies**: Restrict webhook network access using Kubernetes Network Policies
4. **Resource Limits**: Set appropriate CPU and memory limits for the webhook deployment
5. **Updates**: Keep the webhook updated to the latest version
6. **Monitoring**: Monitor webhook logs for suspicious activity

## Security Features

- Request body size limits to prevent memory exhaustion
- No sensitive data in logs
- Minimal container image (FROM scratch)
- Non-root user execution (nobody:65534)
- Separate health/metrics endpoints from webhook API
