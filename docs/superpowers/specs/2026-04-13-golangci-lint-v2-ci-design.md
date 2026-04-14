# golangci-lint v2 CI Unification Design

## Goal

Fix the current GitHub CI failure and unify the repository's lint toolchain on `golangci-lint` v2 so CI, local installs, and the checked-in configuration all use the same version line.

## Context

The latest failed `CI` run exposed two separate problems:

1. `.golangci.yml` declares config `version: 2`, but `.github/workflows/ci.yml` still uses `golangci/golangci-lint-action@v4`, which installs a v1 binary and fails schema validation.
2. `test/integration/integration_test.go` is not `gofmt`-clean, so the separate formatting check fails even if lint is fixed.

This design fixes both issues without expanding into unrelated workflow or documentation cleanup.

## Scope

### In scope

- Update `.github/workflows/ci.yml` so the lint job runs a v2-compatible GitHub Action and a pinned `golangci-lint` v2 release.
- Convert `.golangci.yml` from mixed v1/v2 syntax to valid v2 schema.
- Update `Makefile` so local `install-linter` installs the same v2 version used by CI.
- Reformat `test/integration/integration_test.go` so the existing `gofmt` check passes.
- Re-run the local commands that represent the failing CI paths.

### Out of scope

- Changing the repository's selected lint rules beyond what is required for v2 schema compatibility.
- Refactoring unrelated GitHub Actions jobs.
- Broad documentation cleanup for lint usage.
- Fixing future failures in `security`, `benchmark`, or `fuzz` jobs unless they are directly caused by this migration.

## Design Summary

Use a single pinned `golangci-lint` v2 version everywhere.

- CI uses `golangci/golangci-lint-action@v8` and a pinned `golangci-lint` v2 release.
- Local installs use the v2 module path and the same pinned version.
- `.golangci.yml` is rewritten to pure v2 schema so the checked-in config matches the binary that executes it.
- `gofmt` remains a separate CI check and the failing integration test file is formatted directly.

This keeps the change focused on version alignment and schema correctness instead of turning this into a larger lint policy redesign.

## Detailed Design

### 1. CI lint execution

Modify `.github/workflows/ci.yml` in the `lint` job:

- Change `golangci/golangci-lint-action@v4` to `golangci/golangci-lint-action@v8`.
- Replace `version: latest` with a pinned v2 release.
- Keep the current timeout argument.

Rationale:

- `@v8` supports `golangci-lint` v2.
- Pinning a concrete version avoids another breakage caused by `latest` drifting under a stable checked-in config.

### 2. golangci-lint configuration

Modify `.golangci.yml` so it is valid v2 config throughout.

Required structure changes:

- Keep `version: "2"`.
- Move `linters-settings` to `linters.settings`.
- Move `issues.exclude-rules` to `linters.exclusions.rules`.
- Move `gofmt` and `goimports` settings out of the linter settings block into `formatters.settings`.
- Enable `gofmt` and `goimports` under `formatters.enable`.
- Convert `goimports.local-prefixes` to the v2 list form.

Non-goals inside this file:

- Do not add new linters.
- Do not remove existing exclusions unless the v2 schema forces it.
- Do not broaden scope into policy cleanup.

### 3. Local developer install path

Modify `Makefile`:

- Update `install-linter` from the v1 install path to `github.com/golangci/golangci-lint/v2/cmd/golangci-lint`.
- Pin the same v2 version used in CI.

Rationale:

- `make install-linter` should produce the same major version and behavior as CI.
- This removes the current state where local installs can silently diverge from the workflow.

### 4. Formatting check

Modify `test/integration/integration_test.go` only by running `gofmt -s`.

Rationale:

- The formatting job is correctly catching a real issue.
- This should remain a separate concern from the lint migration.
- The file should be reformatted without semantic changes.

## Execution Flow After the Change

1. CI checks out the repo and installs Go.
2. The lint job invokes `golangci-lint-action@v8`.
3. The action installs the pinned `golangci-lint` v2 binary.
4. The binary loads a fully v2-compatible `.golangci.yml`.
5. Local developers using `make install-linter` install the same v2 version line.
6. The separate formatting step continues to validate `gofmt -s` cleanliness.

This means lint schema failures and format failures stay independently diagnosable.

## Error Handling Strategy

- If v2 reports schema errors, fix only the invalid config structure or value shape.
- If v2 surfaces new lint findings caused by execution differences, handle only migration-related issues required to restore the existing workflow contract.
- If `gofmt` changes `test/integration/integration_test.go`, accept only formatting changes and avoid semantic edits in that file as part of this work.

## Validation Plan

Run these commands locally after the change:

```bash
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4
golangci-lint run --timeout=5m
gofmt -s -l .
go mod verify
go test -v -race -coverprofile=coverage.out -covermode=atomic $(go list ./... | grep -v -e integration -e e2e)
go build -o bin/redmine ./cmd
go test -v -count=1 ./test/e2e/...
```

## Success Criteria

- `.github/workflows/ci.yml` and `Makefile` both target the same pinned `golangci-lint` v2 version.
- `.golangci.yml` contains valid v2 schema and no longer mixes v1 and v2 layout.
- `test/integration/integration_test.go` no longer fails the `gofmt` check.
- The local validation commands that map to the failing CI paths pass.
- A follow-up GitHub `CI` run no longer fails in `Lint` or `Check code formatting` for the reasons seen in the current broken run.

## Risks and Mitigations

- Risk: v2 may expose a small number of new findings.
  - Mitigation: keep the change narrowly focused and fix only migration-related findings required for parity.
- Risk: `latest`-style drift could reintroduce breakage later.
  - Mitigation: pin the action-supported binary version in CI and local install docs.
