# external-dns-unifios-webhook

Webhook provider for [external-dns](https://github.com/kubernetes-sigs/external-dns) that integrates with UniFi OS DNS management. Enables automatic DNS record management in UniFi controllers (UniFi Dream Machine, Cloud Key, etc.) from Kubernetes.

## Features

<div class="grid cards" markdown>

-   :material-dns:{ .lg .middle } **DNS Record Management**

    ---

    Supports A, AAAA, CNAME, and TXT record types with automatic synchronization from Kubernetes resources.

-   :material-filter:{ .lg .middle } **Domain Filtering**

    ---

    Flexible domain filtering with include/exclude patterns and regex support for precise control.

-   :material-heart-pulse:{ .lg .middle } **Health Monitoring**

    ---

    Kubernetes-ready health checks and readiness probes for reliable operation.

-   :material-chart-line:{ .lg .middle } **Prometheus Metrics**

    ---

    Built-in metrics endpoint for monitoring DNS operations and performance.

-   :material-text-box:{ .lg .middle } **Structured Logging**

    ---

    JSON logging with configurable levels for easy integration with log aggregation systems.

-   :material-package-variant-closed:{ .lg .middle } **Minimal Footprint**

    ---

    Lightweight container image built from scratch for security and efficiency.

</div>

## Quick Links

<div class="grid cards" markdown>

-   :material-rocket-launch:{ .lg .middle } **Getting Started**

    ---

    Install and configure the webhook in minutes.

    [:octicons-arrow-right-24: Quick Start](getting-started/quickstart.md)

-   :material-cog:{ .lg .middle } **Configuration**

    ---

    Environment variables and UniFi API setup.

    [:octicons-arrow-right-24: Configuration](configuration/index.md)

-   :material-book-open-variant:{ .lg .middle } **Guides**

    ---

    Integration guides and troubleshooting.

    [:octicons-arrow-right-24: Guides](guides/index.md)

-   :material-code-tags:{ .lg .middle } **Development**

    ---

    Contributing and architecture documentation.

    [:octicons-arrow-right-24: Development](development/index.md)

</div>

## Requirements

| Component | Version | Notes |
|-----------|---------|-------|
| external-dns | v0.20.0+ | Webhook provider support required |
| UniFi Controller | Site API v2 | Any modern UniFi controller |

## Compatibility

Works with any UniFi controller that supports Site API v2, including UniFi Dream Machine series, Cloud Key, and self-hosted UniFi Network Application.

**Tested with:** UniFi OS 4.3.9, UniFi Network 9.4.19

## License

This project is licensed under the [BSD-3-Clause License](https://github.com/lexfrei/external-dns-unifios-webhook/blob/master/LICENSE).
