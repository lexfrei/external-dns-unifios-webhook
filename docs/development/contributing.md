# Contributing

Guidelines for contributing to external-dns-unifios-webhook.

## Getting Started

1. Fork the repository
2. Clone your fork
3. Create a feature branch
4. Make your changes
5. Submit a pull request

## Development Workflow

### Branch Naming

Use descriptive branch names:

- `feat/add-mx-support` - New features
- `fix/connection-timeout` - Bug fixes
- `docs/update-installation` - Documentation
- `refactor/provider-cleanup` - Refactoring

### Commit Messages

Follow semantic commit format:

```
type(scope): description

Optional body explaining what and why.

Co-Authored-By: Your Name <email@example.com>
```

**Types:**

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation |
| `refactor` | Code refactoring |
| `test` | Adding tests |
| `chore` | Maintenance |
| `ci` | CI/CD changes |
| `perf` | Performance |

**Examples:**

```
feat(provider): add MX record support

Implement MX record creation and deletion in the UniFi provider.
```

```
fix(config): handle empty API key gracefully

Return clear error message when API key is not configured.
```

## Code Style

### Go

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Run `golangci-lint run` before committing
- All exported types/functions must have documentation

### Testing

- Write tests for new functionality
- Maintain or improve code coverage
- Use table-driven tests where appropriate

```go
func TestSomething(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid input", "foo", "bar", false},
        {"empty input", "", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Something(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            }
            if got != tt.want {
                t.Errorf("got = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Pull Request Process

### Before Submitting

- [ ] Tests pass locally (`go test ./...`)
- [ ] Linter passes (`golangci-lint run`)
- [ ] Code is documented
- [ ] Commit messages follow format

### PR Description

Use the PR template and include:

1. **Summary**: What does this PR do?
2. **Changes**: Key changes (high-level)
3. **Testing**: How was it tested?

### Review Process

1. Maintainer reviews the PR
2. Address feedback with additional commits
3. Once approved, maintainer merges

## Reporting Issues

### Bug Reports

Include:

- Webhook version
- UniFi OS and Network versions
- Steps to reproduce
- Expected vs actual behavior
- Logs (with API keys redacted)

### Feature Requests

Include:

- Use case description
- Proposed solution
- Alternatives considered

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow

## License

By contributing, you agree that your contributions will be licensed under the BSD-3-Clause License.

## Questions?

- Open a [GitHub Discussion](https://github.com/lexfrei/external-dns-unifios-webhook/discussions)
- Check existing [Issues](https://github.com/lexfrei/external-dns-unifios-webhook/issues)
