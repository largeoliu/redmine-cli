# API Key Masking Design

## Overview

Mask API key input during login so it displays as `*` characters instead of plaintext.

## Problem

Currently `promptSecret` uses `bufio.Reader.ReadString` which shows input in plaintext:
```
API Key: mysecretapikey123
```

This is a security concern as others can see the API key on screen.

## Solution

Use `golang.org/x/term.ReadPassword` to read without terminal echo:
```
API Key: *************
```

## Changes

**File**: `internal/app/login.go`

### 1. Add import
```go
"golang.org/x/term"
```

### 2. Modify promptSecret function

Replace:
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

With:
```go
func promptSecret(reader *bufio.Reader, prompt string) string {
    if prompt != "" {
        fmt.Print(prompt + ": ")
    }
    // ReadPassword reads from terminal without echo, returns '*' mask by default
    oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
    if err != nil {
        // Fallback: try non-raw mode
        input, _ := reader.ReadString('\n')
        return strings.TrimSpace(input)
    }
    defer term.Restore(int(os.Stdin.Fd()), oldState)

    input, err := term.ReadPassword(int(os.Stdin.Fd()))
    fmt.Println() // newline after masked input
    if err != nil {
        return ""
    }
    return strings.TrimSpace(string(input))
}
```

**Note**: The `reader` parameter is kept for fallback compatibility but not used in the primary path.

## Expected Behavior

```
$ redmine login
Redmine URL: https://redmine.example.com
正在检查连通性... ✓
API Key: ********        <- input hidden, shows as *
正在验证连接... ✓
...
```

## Platform Support

- Linux: via `golang.org/x/term` (wraps `termios`)
- macOS: via `golang.org/x/term` (wraps `termios`)
- Windows: via `golang.org/x/term` (wraps `consolev2`)

## Scope

- Only affects `promptSecret` used for API Key input
- `promptInput` remains unchanged (plaintext display for non-sensitive input)
- Other commands unaffected
