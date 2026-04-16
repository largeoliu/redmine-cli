# API Key Masking Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Mask API key input with `*` characters during login instead of showing plaintext

**Architecture:** Use `golang.org/x/term.ReadPassword` to read from terminal without echo. Falls back to plaintext input on non-terminal environments.

**Tech Stack:** `golang.org/x/term` (Go stdlib extension)

---

## File Structure

- Modify: `internal/app/login.go` (promptSecret function)
- Modify: `go.mod` (add dependency)

---

### Task 1: Add golang.org/x/term dependency

**Files:**
- Modify: `go.mod`

- [ ] **Step 1: Add dependency**

```bash
go get golang.org/x/term
```

- [ ] **Step 2: Verify dependency added**

Run: `grep "golang.org/x/term" go.mod`
Expected: `golang.org/x/term v0.x.x`

---

### Task 2: Modify promptSecret to use term.ReadPassword

**Files:**
- Modify: `internal/app/login.go:120-129`

**Original code (lines 120-129):**
```go
func promptSecret(reader *bufio.Reader, prompt string) string {
    if prompt != "" {
        fmt.Print(prompt + ": ")
    }
    input, err := reader.ReadString('\n')
    if err != nil {
        return ""
    }
    return strings.TrimSpace(input)
}
```

- [ ] **Step 1: Add golang.org/x/term import**

Add to import block in `internal/app/login.go`:
```go
"golang.org/x/term"
```

- [ ] **Step 2: Replace promptSecret function**

Replace the original `promptSecret` function with:

```go
func promptSecret(reader *bufio.Reader, prompt string) string {
    if prompt != "" {
        fmt.Print(prompt + ": ")
    }

    // Try to use terminal password mode for masked input
    oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
    if err != nil {
        // Fallback: not a terminal, use plain input
        input, _ := reader.ReadString('\n')
        return strings.TrimSpace(input)
    }
    defer term.Restore(int(os.Stdin.Fd()), oldState)

    // ReadPassword reads without echo, displays '*' mask by default
    input, err := term.ReadPassword(int(os.Stdin.Fd()))
    fmt.Println() // newline after masked input
    if err != nil {
        return ""
    }
    return strings.TrimSpace(string(input))
}
```

- [ ] **Step 3: Verify the file compiles**

Run: `go build ./internal/app/`
Expected: No errors

---

### Task 3: Verify login command still works

**Files:**
- Test: `internal/app/login_test.go`

- [ ] **Step 1: Run existing login tests**

Run: `go test ./internal/app/ -run Login -v`
Expected: All tests pass

---

### Task 4: Commit changes

- [ ] **Step 1: Stage and commit**

```bash
git add go.mod go.sum internal/app/login.go
git commit -m "feat(login): mask API key input with '*' for security"
```

---

### Task 5: Add boundary tests for promptSecret

**Files:**
- Modify: `internal/app/login_test.go`

**Added tests:**
- `TestPromptSecretBoundaryCases`: empty input, whitespace-only, special chars, long input
- `TestPromptSecretFallbackMode`: verifies fallback path works when term.MakeRaw fails

- [ ] **Step 1: Run boundary tests**

Run: `go test ./internal/app/ -run "PromptSecret" -v`
Expected: All tests pass

---

## Verification

After implementation, run `redmine login` and confirm:
1. When prompted for "API Key:", typing shows `*` characters
2. After pressing Enter, the key is still correctly stored/validated
