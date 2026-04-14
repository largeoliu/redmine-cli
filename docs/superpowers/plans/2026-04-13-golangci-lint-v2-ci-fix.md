# golangci-lint v2 CI Fix Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Align GitHub CI, local lint installation, and the checked-in lint configuration on `golangci-lint` `v2.11.4`, and clear the current `gofmt` failure in `test/integration/integration_test.go`.

**Architecture:** Treat this as one toolchain-alignment change across four files: the GitHub Action picks the binary, `.golangci.yml` defines the schema and enabled formatters, `Makefile` installs the local binary, and `test/integration/integration_test.go` must be formatting-clean. Keep the fix narrow: pin versions, migrate the config to valid v2 syntax, format the single failing Go file, then run the repo's existing CI-equivalent verification commands.

**Tech Stack:** GitHub Actions YAML, `golangci-lint` `v2.11.4`, Go `1.23`, `gofmt`, Make

---

## File Structure

- Modify: `.github/workflows/ci.yml` - pin the lint job to `golangci/golangci-lint-action@v8` and `golangci-lint` `v2.11.4`.
- Modify: `.golangci.yml` - migrate the mixed v1/v2 layout to pure v2 schema and move formatter settings under `formatters`.
- Modify: `Makefile` - make `install-linter` install the same `golangci-lint` `v2.11.4` binary that CI uses.
- Modify: `test/integration/integration_test.go` - apply formatting-only cleanup with `gofmt -s`.
- Reference: `docs/superpowers/specs/2026-04-13-golangci-lint-v2-ci-design.md` - approved design source for scope and success criteria.

### Task 1: Pin the lint toolchain version in CI and local install paths

**Files:**
- Modify: `.github/workflows/ci.yml`
- Modify: `Makefile`
- Reference: `docs/superpowers/specs/2026-04-13-golangci-lint-v2-ci-design.md`

- [ ] **Step 1: Reproduce the current CI root cause from the latest failed run**

Run: `gh run view 24333691700 --log-failed`
Expected: the log shows `golangci/golangci-lint-action@v4` installing a v1 binary and failing with `you are using a configuration file for golangci-lint v2 with golangci-lint v1`.

- [ ] **Step 2: Update the lint action in `.github/workflows/ci.yml`**

Replace the lint step with this exact block:

```yaml
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v8
        with:
          version: v2.11.4
          args: --timeout=5m
```

- [ ] **Step 3: Update `Makefile` so local installs use the same pinned v2 binary**

Replace the `install-linter` target with this exact block:

```make
install-linter:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4
```

- [ ] **Step 4: Verify both files now point at the same v2 version line**

Run: `rg -n "golangci-lint-action@|v2\.11\.4|golangci-lint/v2" .github/workflows/ci.yml Makefile`
Expected: `.github/workflows/ci.yml` shows `golangci-lint-action@v8` and `version: v2.11.4`; `Makefile` shows `github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4`.

- [ ] **Step 5: If the user explicitly requests a commit, create the toolchain-alignment commit**

```bash
git add .github/workflows/ci.yml Makefile
git commit -m "ci: pin golangci-lint to v2.11.4"
```

### Task 2: Migrate `.golangci.yml` to valid v2 schema

**Files:**
- Modify: `.golangci.yml`
- Test: `.golangci.yml`

- [ ] **Step 1: Install the pinned v2 binary and reproduce the current config failure before editing**

Run: `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4 && golangci-lint run --timeout=5m`
Expected: FAIL with config validation errors caused by the current mixed schema, referencing old keys such as `linters-settings` or `issues.exclude-rules`.

- [ ] **Step 2: Replace `.golangci.yml` with this exact v2 config**

```yaml
version: "2"

linters:
  enable:
    - errcheck
    - govet
    - ineffassign
    - staticcheck
    - unused
    - gosec
    - gocyclo
    - dupl
    - misspell
    - revive
    - prealloc
  settings:
    errcheck:
      check-type-assertions: true
      check-blank: true
    govet:
      enable-all: true
      disable:
        - fieldalignment
    gocyclo:
      min-complexity: 15
    dupl:
      threshold: 100
    misspell:
      locale: US
    revive:
      rules:
        - name: blank-imports
        - name: context-as-argument
        - name: context-keys-type
        - name: dot-imports
        - name: error-return
        - name: error-strings
        - name: error-naming
        - name: exported
        - name: if-return
        - name: increment-decrement
        - name: var-naming
        - name: var-declaration
        - name: range
        - name: receiver-naming
        - name: time-naming
        - name: unexported-return
        - name: indent-error-flow
        - name: errorf
        - name: empty-block
        - name: superfluous-else
        - name: unused-parameter
        - name: unreachable-code
        - name: redefines-builtin-id
    gosec:
      excludes:
        - G104
  exclusions:
    rules:
      - path: _test\.go
        linters:
          - dupl
          - gosec
          - errcheck
          - gocyclo
          - revive
          - unused
          - staticcheck
          - misspell
          - prealloc
          - govet
          - ineffassign
      - path: (^|/)test/e2e/
        linters:
          - dupl
          - gosec
          - errcheck
          - gocyclo
          - revive
          - unused
          - staticcheck
          - misspell
          - prealloc
          - govet
          - ineffassign
      - path: (^|/)test/integration/
        linters:
          - dupl
          - gosec
          - errcheck
          - gocyclo
          - revive
          - unused
          - staticcheck
          - misspell
          - prealloc
          - govet
          - ineffassign

formatters:
  enable:
    - gofmt
    - goimports
  settings:
    gofmt:
      simplify: true
    goimports:
      local-prefixes:
        - github.com/largeoliu/redmine-cli

issues:
  max-issues-per-linter: 0
  max-same-issues: 0
```

- [ ] **Step 3: Re-run lint and confirm the schema error is gone**

Run: `golangci-lint run --timeout=5m`
Expected: the command no longer fails with v1/v2 schema-validation errors; if any failure remains at this point, it should be a source-level issue such as formatting, not config parsing.

- [ ] **Step 4: If the user explicitly requests a commit, create the config-migration commit**

```bash
git add .golangci.yml
git commit -m "ci: migrate golangci-lint config to v2"
```

### Task 3: Clear the standalone `gofmt` failure

**Files:**
- Modify: `test/integration/integration_test.go`
- Test: `test/integration/integration_test.go`

- [ ] **Step 1: Reproduce the current formatting failure before changing the file**

Run: `gofmt -s -l .`
Expected: output includes exactly `test/integration/integration_test.go`.

- [ ] **Step 2: Apply formatting-only cleanup to the integration test file**

Run:

```bash
gofmt -s -w test/integration/integration_test.go
```

Expected: no terminal output; the file is rewritten in place with formatting-only changes.

- [ ] **Step 3: Verify the repository is now formatting-clean**

Run: `gofmt -s -l .`
Expected: no output.

- [ ] **Step 4: If the user explicitly requests a commit, create the formatting commit**

```bash
git add test/integration/integration_test.go
git commit -m "style: format integration test file"
```

### Task 4: Run CI-equivalent verification and capture the final state

**Files:**
- Test: `.github/workflows/ci.yml`
- Test: `.golangci.yml`
- Test: `Makefile`
- Test: `test/integration/integration_test.go`

- [ ] **Step 1: Run the pinned v2 linter end-to-end**

Run: `golangci-lint run --timeout=5m`
Expected: PASS with exit code 0.

- [ ] **Step 2: Verify Go module integrity**

Run: `go mod verify`
Expected: `all modules verified`.

- [ ] **Step 3: Run the same unit-test coverage command used by CI**

Run: `go test -v -race -coverprofile=coverage.out -covermode=atomic $(go list ./... | grep -v -e integration -e e2e)`
Expected: PASS and writes `coverage.out`.

- [ ] **Step 4: Build the CLI binary exactly as CI does for E2E setup**

Run: `go build -o bin/redmine ./cmd`
Expected: PASS with exit code 0 and writes `bin/redmine`.

- [ ] **Step 5: Run the E2E suite**

Run: `go test -v -count=1 ./test/e2e/...`
Expected: PASS.

- [ ] **Step 6: Inspect the final diff and working tree before handing back to the user**

Run: `git status --short && git diff --stat`
Expected: only `.github/workflows/ci.yml`, `.golangci.yml`, `Makefile`, `test/integration/integration_test.go`, and the saved plan/spec artifacts appear as intentional changes.

- [ ] **Step 7: If the user explicitly requests a final commit, create it after all verification passes**

```bash
git add .github/workflows/ci.yml .golangci.yml Makefile test/integration/integration_test.go
git commit -m "ci: align golangci-lint with v2 config"
```
