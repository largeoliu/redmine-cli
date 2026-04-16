# Contributing to redmine-cli

Thank you for your interest in contributing! This guide covers the basics.

## Setup

1. Clone the repository
2. Install Go 1.23+
3. Install golangci-lint: `make install-linter`
4. Install git hooks: `make install-hooks`

## Development

### Build

```sh
make build
```

### Test

```sh
make test            # Unit tests
make e2e             # E2E tests
make test-coverage   # Coverage report
```

### Lint

```sh
make lint       # Check
make lint-fix   # Auto-fix
```

### Integration Tests

Integration tests require a running Redmine instance:

```sh
make integration-setup  # Create .env file
make integration-with-env # Run with .env
```

## Pull Request Process

1. Create a branch from `main`
2. Make your changes with tests
3. Ensure all checks pass: `make lint && make test`
4. Submit a PR with a clear description

## Code Style

- Follow `gofmt` and `goimports` formatting
- All linter rules in `.golangci.yml` must pass
- Write tests for new functionality
- Keep coverage above 90%
