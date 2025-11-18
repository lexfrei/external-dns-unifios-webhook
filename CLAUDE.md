# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an external-dns webhook provider for UniFi OS DNS management. It enables automatic DNS record management in UniFi controllers (Dream Machine, Cloud Key, etc.) from Kubernetes using external-dns.

**Key technologies:**
- Go 1.25.4
- OpenAPI code generation (oapi-codegen)
- UniFi API client: github.com/lexfrei/go-unifi
- External DNS: sigs.k8s.io/external-dns
- Chi router for HTTP handling
- Structured logging with slog
- Prometheus metrics

## Build and Development Commands

### Building

```bash
# Build the webhook binary
go build -o webhook ./cmd/webhook

# Build with version information (used in CI)
go build -ldflags "-s -w -X main.Version=v1.0.0 -X main.Gitsha=abc123" -trimpath ./cmd/webhook
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with race detection and coverage (CI standard)
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# Run tests for specific package
go test ./internal/provider/...

# Run specific test
go test -v -run TestUniFiProvider_Records ./internal/provider/
```

### Linting

```bash
# Run golangci-lint (must pass with zero errors before any push)
golangci-lint run

# Run with timeout for slower machines
golangci-lint run --timeout=5m

# Auto-fix issues where possible
golangci-lint run --fix
```

**CRITICAL:** ALL linting errors must be fixed before pushing. See `.golangci.yaml` for configuration. Notable settings:
- Function length limit: 60 lines/statements (funlen)
- Cyclomatic complexity: 15 (gocyclo, cyclop)
- Minimum variable name length: 3 characters (varnamelen)
- All linters enabled by default except: depguard, exhaustruct, gochecknoinits, nonamedreturns, wsl, lll, godoclint, errchkjson

### Running Locally

```bash
# Set required environment variables
export WEBHOOK_UNIFI_HOST="https://192.168.1.1"
export WEBHOOK_UNIFI_API_KEY="your-api-key"
export WEBHOOK_LOGGING_LEVEL="debug"

# Run the webhook server
go run ./cmd/webhook

# The webhook listens on:
# - localhost:8888 - webhook API endpoints
# - 0.0.0.0:8080 - health/metrics endpoints (/healthz, /readyz, /metrics)
```

### Container Operations

```bash
# Build container image (use podman, not docker)
podman build --tag external-dns-unifios-webhook:local --file Containerfile .

# Run container locally
podman run --rm \
  --env WEBHOOK_UNIFI_HOST="https://192.168.1.1" \
  --env WEBHOOK_UNIFI_API_KEY="your-api-key" \
  --publish 8888:8888 \
  --publish 8080:8080 \
  external-dns-unifios-webhook:local

# If "Cannot connect to Podman" error occurs (macOS/Windows)
podman machine start
```

## Git Workflow and Commit Standards

### Commit Message Format

Use **Semantic Commit Messages** with Claude attribution:

**Format:**

```text
type(scope): brief description of changes

Optional longer explanation of what was changed and why.

Co-Authored-By: Claude <noreply@anthropic.com>
```

**IMPORTANT**: Do NOT include "🤖 Generated with [Claude Code]" anywhere. The `Co-Authored-By: Claude <noreply@anthropic.com>` line is sufficient attribution for commits only. In PR descriptions, comments, documentation, and all other content - no Claude attribution is needed at all.

**Types:**

- `feat`: New features
- `fix`: Bug fixes
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding/updating tests
- `chore`: Maintenance tasks
- `ci`: CI/CD pipeline changes
- `perf`: Performance improvements
- `build`: Build system changes

**Level of Detail:**

- **Avoid excessive technical details** - the diff shows WHAT changed, commit/PR should explain WHY
- Focus on high-level changes and their purpose, not implementation specifics
- Bad example: "Modified jq command to output `-t` and tag on separate lines using `\"-t\", .` syntax"
- Good example: "Fix bash array construction in manifest creation"
- Bad example: "Replaced `find . -type f` with glob pattern `*` for cleaner digest file iteration"
- Good example: "Improve digest file handling"
- Excessive technical details increase cognitive load and obscure the actual purpose
- Code review shows implementation details - commit message should explain the change rationale

### Pull Request Standards

When creating Pull Requests, follow these strict guidelines:

1. **ALWAYS create PR in DRAFT mode by default**
   - Use `gh pr create --draft` flag to create draft PR
   - **ALWAYS show PR text to user BEFORE creating** - ask for approval
   - **ALL PR content MUST be in English** (title, description, all text)
   - Exception: User explicitly requests non-draft PR

2. **Search for PR template**
   - Before creating PR: search `.github/` directory for pull request templates
   - Templates may be: `pull_request_template.md`, `PULL_REQUEST_TEMPLATE.md`, or in `.github/PULL_REQUEST_TEMPLATE/`

3. **Verify template requirements**
   - Ensure ALL template requirements are actually fulfilled (tests, linters, documentation, etc.)
   - Do NOT check boxes that are not truly completed
   - If requirements cannot be met, explain why in PR description

4. **Create PR body from template**
   - Use the complete template structure
   - Do NOT remove sections from the template
   - Fill ALL sections completely and accurately
   - Keep all checkboxes and checklists from template

5. **PR Title format**
   - MUST follow semantic commit format: `type(scope): title`
   - Examples:
     - `feat(api): add user authentication endpoint`
     - `fix(ui): correct button alignment on mobile`
     - `ci(workflows): optimize container builds with native ARM64 runners`
   - Scope should be specific and meaningful
   - Title should be concise and descriptive

6. **PR Body content guidelines**
   - Do NOT mention specific commit hashes or commit messages
   - Focus on WHAT changed and WHY, not HOW (commits show HOW)
   - **Avoid excessive technical details** - diff shows implementation, PR explains purpose
   - Describe changes at a high level, not line-by-line code changes
   - Example: Instead of "Modified jq to output -t and tag separately", write "Fix manifest creation"
   - Avoid specific performance numbers unless essential (e.g., "A is 81% better than B" - too specific!)
   - Use general terms: "significantly faster", "improved performance", "reduced build time"
   - Be technical and factual, avoid marketing language
   - Exception: Specific numbers are OK for breaking changes, API changes, or when precision matters

7. **Technical accuracy**
   - Describe changes accurately and completely
   - Include all significant modifications
   - Mention breaking changes explicitly
   - Reference related issues if applicable

## Architecture

### Entry Point
- `cmd/webhook/main.go` - Application entry point, server initialization, graceful shutdown

### Core Components

**Provider Layer** (`internal/provider/`)
- `provider.go` - Core UniFi DNS provider implementation
- `interface.go` - Provider interface definition for dependency injection
- Implements external-dns provider interface: `Records()`, `ApplyChanges()`, `AdjustEndpoints()`, `GetDomainFilter()`
- Uses parallel operations with semaphore (max 5 concurrent operations) to optimize API calls
- Record index caching to avoid N*API_calls problem during batch operations

**Configuration** (`internal/config/`)
- `config.go` - Viper-based configuration with environment variable binding
- All env vars prefixed with `WEBHOOK_` (e.g., `WEBHOOK_UNIFI_HOST`)
- Supports nested configuration via dots: `unifi.api_key` → `WEBHOOK_UNIFI_API_KEY`

**HTTP Servers** (`internal/webhookserver/`, `internal/healthserver/`)
- Separate servers for webhook API (localhost:8888) and health/metrics (0.0.0.0:8080)
- OpenAPI-generated handlers from `api/webhook/openapi.yaml` and `api/health/openapi.yaml`
- Generated code in `api/*/generated.go` (do NOT edit manually)

**Observability** (`internal/observability/`, `internal/metrics/`)
- `slog_adapter.go` - Adapts slog to UniFi client logger interface
- `prometheus_recorder.go` - Adapts Prometheus to UniFi client metrics interface
- `metrics.go` - Custom Prometheus metrics for DNS operations
- Metrics: `external_dns_unifi_records_managed`, `external_dns_unifi_operations_total`, `external_dns_unifi_operation_duration_seconds`, `external_dns_unifi_changes_applied`

### Important Implementation Details

**Performance Optimization:**
- Provider uses `buildRecordIndex()` to create in-memory map of DNS records by name before parallel operations
- This prevents N*API_calls problem: without index, each deletion would call ListDNSRecords independently
- With index: 1 ListDNSRecords call + N parallel deletes (10+ records: 2-5s → 200-400ms)

**UniFi API Specifics:**
- TXT records do NOT support TTL field (UniFi API limitation)
- Creates separate DNS records for multi-target endpoints (enables round-robin DNS)
- Default TTL: 300 seconds

**Record Type Mapping:**
- Supports: A, AAAA, CNAME, MX, NS, SRV, TXT
- Maps between external-dns types and UniFi API types

**Error Handling:**
- Uses cockroachdb/errors for error wrapping with stack traces
- Parallel operations collect errors in channel and aggregate them

**External-DNS Retry Behavior:**
- External-dns ONLY retries on 5xx (server errors) and 429 (Too Many Requests)
- 4xx errors (client errors) are NOT retried - considered permanent failures
- This includes:
  - 400 Bad Request - malformed data, will not retry
  - 413 Request Entity Too Large - request too big, will not retry
  - 404 Not Found - endpoint missing, will not retry
- Request body limit: 5MB (supports ~25,000 DNS records per request)
- Production observation: deployments with 20,000+ DNS records exist
- If webhook returns 5xx on legitimate requests, external-dns will retry indefinitely

## Code Generation

**OpenAPI code generation:**
```bash
# API definitions are in api/*/openapi.yaml
# Generated code goes to api/*/generated.go
# Regenerate when OpenAPI specs change (typically done by upstream external-dns)

# No explicit generation command in repo - likely uses go:generate directives or manual oapi-codegen
```

## Configuration

**Required environment variables:**
- `WEBHOOK_UNIFI_HOST` - UniFi controller URL (must be IP, not hostname like unifi.local)
- `WEBHOOK_UNIFI_API_KEY` - API key from UniFi controller

**Optional environment variables:**
- `WEBHOOK_UNIFI_SITE` - UniFi site name (default: "default")
- `WEBHOOK_UNIFI_SKIP_TLS_VERIFY` - Skip TLS verification (default: true)
- `WEBHOOK_SERVER_HOST` - Webhook bind address (default: "localhost")
- `WEBHOOK_SERVER_PORT` - Webhook port (default: "8888")
- `WEBHOOK_HEALTH_HOST` - Health server bind address (default: "0.0.0.0")
- `WEBHOOK_HEALTH_PORT` - Health server port (default: "8080")
- `WEBHOOK_LOGGING_LEVEL` - Log level: debug, info, warn, error (default: "info")
- `WEBHOOK_LOGGING_FORMAT` - Log format: json, text (default: "json")
- `WEBHOOK_DEBUG_PPROF_ENABLED` - Enable pprof profiling (default: false, DO NOT use in production)
- `WEBHOOK_DEBUG_PPROF_PORT` - pprof port (default: "6060")

## CI/CD Workflows

**PR Workflow** (`.github/workflows/pr.yaml`):
1. Lint with golangci-lint
2. Test with race detection and coverage
3. Build multi-arch containers (linux/amd64, linux/arm64) using native ARM64 runners
4. Push PR image as `ghcr.io/lexfrei/external-dns-unifios-webhook:pr-<number>`
5. Cleanup PR image when PR is closed

**Release Workflow** (`.github/workflows/release.yaml`):
- Triggered on version tags
- Builds and publishes release containers

## Testing Strategy

**Test packages:**
- `internal/provider/provider_test.go` - Provider unit tests with mock UniFi client
- `internal/healthserver/server_test.go` - Health server tests with readiness cache

**Test conventions:**
- Use table-driven tests with `tt` variable for test case
- Mock external dependencies (UniFi API client) via interfaces
- Same-package tests allowed when testing private functions/types (see `//nolint:testpackage` comments)

## Dependencies

**Key external dependencies:**
- `github.com/lexfrei/go-unifi/api/network` - UniFi API client (custom library)
- `sigs.k8s.io/external-dns` - External DNS types and interfaces
- `github.com/go-chi/chi/v5` - HTTP router
- `github.com/go-chi/httplog/v3` - HTTP request logging
- `github.com/prometheus/client_golang` - Prometheus metrics
- `github.com/spf13/viper` - Configuration management
- `github.com/cockroachdb/errors` - Enhanced error handling
- `golang.org/x/sync/semaphore` - Concurrency limiting

## Common Development Patterns

**Adding new DNS record type:**
1. Add mapping in `unifiToEndpoint()` (UniFi → endpoint)
2. Add mapping in `endpointToUniFiWithTarget()` (endpoint → UniFi)
3. Check UniFi API constraints (e.g., TXT records have no TTL)
4. Add test cases in `provider_test.go`

**Adding new metrics:**
1. Define metric in `internal/metrics/metrics.go`
2. Register in `Register()` function
3. Update metric in appropriate provider method
4. Test metric presence in health endpoint

**Adding new configuration:**
1. Add field to config struct in `internal/config/config.go`
2. Add `BindEnv()` call for environment variable
3. Add default in `setDefaults()`
4. Add validation in `validate()` if required
5. Document in README.md

## Container Image

**Build process:**
- Multi-stage build using golang:1.25-alpine
- Binary compressed with UPX (--best --lzma)
- Final image FROM scratch (minimal attack surface)
- Runs as nobody (65534) for security
- Includes only binary and CA certificates

**Image location:**
- Production: `ghcr.io/lexfrei/external-dns-unifios-webhook:latest`
- PR images: `ghcr.io/lexfrei/external-dns-unifios-webhook:pr-<number>`
- Release tags: `ghcr.io/lexfrei/external-dns-unifios-webhook:v1.0.0`

## Release Process

### Versioning Strategy

Follow Semantic Versioning (semver) with pre-1.0 conventions:

**Version format: `v{MAJOR}.{MINOR}.{PATCH}`**

- **0.0.x → 0.0.(x+1)** - Patch releases (bug fixes, minor improvements)
- **0.0.x → 0.(y+1).0** - Minor releases (new features, significant improvements, dependency updates)
- **0.x.y → (x+1).0.0** - Major releases (breaking changes, significant API changes)

**Version decision criteria:**
- Bug fixes only → PATCH bump
- New features, performance improvements, or dependency updates → MINOR bump
- Breaking changes or major API changes → MAJOR bump

### Release Notes Style

Release notes must follow the established project format for consistency.

**Structure:**
```markdown
Brief introduction (1-2 sentences about the release focus)

## What's New

**Feature/Improvement Category 1**
- Bullet point describing improvement
- Another bullet point with details
- Focus on user-visible benefits

**Feature/Improvement Category 2**
- Similar format
- Group related changes together

**Dependency Updates**
- List dependency version changes
- Include key improvements from updated dependencies

## Breaking Changes
None. (or list specific breaking changes)

**Full Changelog**: https://github.com/lexfrei/external-dns-unifios-webhook/compare/vX.Y.Z...vA.B.C
```

**Style guidelines:**
- **Introduction**: One brief sentence about the release theme/focus
- **Category headers**: Bold, descriptive (e.g., "Performance Optimizations", "Memory Leak Fix", "Enhanced Capacity")
- **Bullet points**: Focus on user-visible impact, not implementation details
- **Technical but accessible**: Explain WHAT changed and WHY, not HOW
- **Group by theme**: Performance, observability, reliability, capacity, etc.
- **Breaking changes section**: Always include, even if "None"
- **Full changelog link**: Always include comparison link to previous version
- **Language**: Professional, factual, avoid marketing language
- **Avoid specifics**: General terms like "significantly faster" instead of "81% faster"

**Example categories used:**
- Memory Leak Fix
- Performance Optimizations
- Enhanced Capacity
- Observability Improvements
- Reliability Improvements
- HTTP Request Logging
- Readiness Probe Optimization
- Batch Operation Performance
- Dependency Updates

### Release Workflow

**Prerequisites:**
- All changes merged to master branch
- All tests passing
- All linters passing
- GPG signing configured

**Steps:**

1. **Determine version number**
   ```bash
   # Check last tag
   git tag --list --sort=-v:refname | head -5

   # Analyze changes since last release
   git log v{LAST_VERSION}..HEAD --oneline
   gh pr view {PR_NUMBER} --json title,body,commits
   ```

2. **Prepare release notes**
   - Review merged PRs and commits
   - Follow release notes style (see above)
   - Group changes by category
   - Focus on user impact
   - Review previous releases for consistency: `gh release view v{VERSION}`

3. **Create signed tag**
   ```bash
   # CRITICAL: ALWAYS use --sign flag, NEVER use --annotate
   git tag --sign v{VERSION} --message "$(cat <<'EOF'
   v{VERSION} - Brief Release Title

   Brief introduction sentence about the release.
   EOF
   )"

   # Verify tag was created
   git tag --list v{VERSION} --format='%(refname:short) - %(subject)'
   ```

4. **Push tag**
   ```bash
   git push origin v{VERSION}
   ```

5. **Wait for CI**
   - Release workflow triggers automatically on tag push
   - Builds multi-arch containers (linux/amd64, linux/arm64)
   - Creates GitHub release with auto-generated notes
   - Monitor: `gh run list --workflow=release.yaml`

6. **Update release notes**
   ```bash
   # After CI creates release, update with prepared release notes
   gh release edit v{VERSION} --notes "$(cat <<'EOF'
   [Paste prepared release notes here]
   EOF
   )"

   # Verify release
   gh release view v{VERSION}
   ```

**Important reminders:**
- ALWAYS sign tags with `--sign` flag (GPG signing required)
- NEVER use `--annotate` or `-a` flags (creates unsigned tags)
- Release notes must be in English (public content requirement)
- Tag message can be brief, detailed notes go in GitHub release
- CI handles container building and initial release creation
- Final step is updating release text with proper formatted notes
