# Contribution Guide

## Branches

- Use `feature/*` for new capability work.
- Use `fix/*` for bug fixes.
- Use `release/*` for release preparation.

## Commits

- Follow Conventional Commits such as `feat:`, `fix:`, `perf:`, `refactor:`, `test:`, and `docs:`.

## Local hooks

1. Run `git config core.hooksPath .githooks`.
2. Make sure `golangci-lint`, `npm`, and the web dependencies are available locally.

The pre-commit hook runs:

- `golangci-lint run ./...`
- `npm --prefix web run lint`
- `npm --prefix web run format:check`

## Integration tests

Use Docker Compose to bring up real MySQL and Redis dependencies:

- Windows: `powershell -ExecutionPolicy Bypass -File scripts/run-integration-tests.ps1`
- Unix-like: `bash scripts/run-integration-tests.sh`

By default the scripts use host ports `3307` for MySQL and `6380` for Redis so they do not clash with local development services. Override `MYSQL_PORT` and `REDIS_PORT` if needed.

The integration suite is gated by `RUN_INTEGRATION_TESTS=1`, so regular `go test ./...` stays fast and local-environment independent.
