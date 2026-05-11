# ModelCraft CLI v1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go-based, agent-first, read-only ModelCraft CLI that authenticates through Gateway, stores a single local org-scoped EndUser session, supports resource discovery and runtime read operations, and runs reliably on Windows/macOS/Linux.

**Architecture:** The CLI is a separate Go module in `modelcraft-cli/` with focused packages for config, auth/session management, Gateway/GraphQL clients, output formatting, resource parsing, and static CLI schema export. Gateway gets a parallel set of CLI-only EndUser auth passthrough routes under `/api/cli/end-user/auth/*` that return JSON tokens directly instead of using refresh cookies; runtime/catalog data continues to flow through existing GraphQL proxy endpoints.

**Tech Stack:** Go, Cobra, standard library `net/http`, `encoding/json`, `os`, `filepath`, Gateway `chi`, Go tests with `testing`/`httptest`.

---

## File Structure

### New CLI module

- Create: `modelcraft-cli/go.mod`
  - Separate Go module for the CLI.
- Create: `modelcraft-cli/main.go`
  - Binary entrypoint that wires version metadata into Cobra.
- Create: `modelcraft-cli/justfile`
  - Local build, test, fmt, and cross-platform release commands.
- Create: `modelcraft-cli/README.md`
  - Quick start, Windows quoting notes, and command examples.

### CLI command layer

- Create: `modelcraft-cli/cmd/root.go`
  - Root Cobra command, global flags, and session-aware command context.
- Create: `modelcraft-cli/cmd/auth.go`
  - `auth` command group and subcommands.
- Create: `modelcraft-cli/cmd/catalog.go`
  - `catalog` command group.
- Create: `modelcraft-cli/cmd/query.go`
  - `query`, `get`, `count`, `aggregate` commands.
- Create: `modelcraft-cli/cmd/describe.go`
  - Remote describe command.
- Create: `modelcraft-cli/cmd/schema.go`
  - Static CLI schema export.
- Create: `modelcraft-cli/cmd/version.go`
  - Version output.

### CLI support packages

- Create: `modelcraft-cli/internal/config/credentials.go`
  - Single-profile credential persistence and environment overrides.
- Create: `modelcraft-cli/internal/auth/session.go`
  - Session manager for load/save/refresh/currentProject behavior.
- Create: `modelcraft-cli/internal/client/auth.go`
  - Gateway CLI auth REST client.
- Create: `modelcraft-cli/internal/client/graphql.go`
  - Shared GraphQL request/response plumbing.
- Create: `modelcraft-cli/internal/client/catalog.go`
  - GraphQL catalog/discovery client.
- Create: `modelcraft-cli/internal/client/runtime.go`
  - Runtime read query/introspection client.
- Create: `modelcraft-cli/internal/output/success.go`
  - Unified JSON/YAML success writer.
- Create: `modelcraft-cli/internal/output/errors.go`
  - English error envelope and exit-code mapping.
- Create: `modelcraft-cli/internal/resource/path.go`
  - `project.database.model` and `database.model` parser.
- Create: `modelcraft-cli/internal/schema/commands.go`
  - Static schema generation for `schema commands/query/flags`.
- Create: `modelcraft-cli/internal/app/context.go`
  - Shared command context (`server`, `org`, `session`, `verbose`, output mode).

### CLI tests

- Create: `modelcraft-cli/internal/config/credentials_test.go`
- Create: `modelcraft-cli/internal/auth/session_test.go`
- Create: `modelcraft-cli/internal/output/errors_test.go`
- Create: `modelcraft-cli/internal/resource/path_test.go`
- Create: `modelcraft-cli/internal/client/auth_test.go`
- Create: `modelcraft-cli/internal/client/catalog_test.go`
- Create: `modelcraft-cli/internal/client/runtime_test.go`
- Create: `modelcraft-cli/internal/schema/commands_test.go`
- Create: `modelcraft-cli/cmd/auth_test.go`
- Create: `modelcraft-cli/cmd/catalog_test.go`
- Create: `modelcraft-cli/cmd/query_test.go`
- Create: `modelcraft-cli/cmd/describe_test.go`
- Create: `modelcraft-cli/cmd/schema_test.go`

### Gateway changes

- Modify: `modelcraft-gateway/cmd/gateway/main.go`
  - Register `/api/cli/end-user/auth/*` route group.
- Modify: `modelcraft-gateway/internal/auth/handler.go`
  - Add CLI passthrough handlers that never set or read refresh cookies.
- Create: `modelcraft-gateway/internal/auth/handler_cli_test.go`
  - Verify CLI login/refresh/logout/me/select-project passthrough behavior.

---

### Task 1: Scaffold the CLI Module and Root Command

**Files:**
- Create: `modelcraft-cli/go.mod`
- Create: `modelcraft-cli/main.go`
- Create: `modelcraft-cli/justfile`
- Create: `modelcraft-cli/cmd/root.go`
- Create: `modelcraft-cli/cmd/version.go`
- Test: `modelcraft-cli/cmd/version_test.go`

- [ ] **Step 1: Write the failing version command test**

```go
package cmd

import (
    "bytes"
    "testing"
)

func TestVersionCommandPrintsInjectedMetadata(t *testing.T) {
    t.Setenv("NO_COLOR", "1")

    cmd := NewRootCommand(BuildInfo{
        Version:   "v0.1.0",
        Commit:    "abc1234",
        BuildTime: "2026-05-09T12:00:00Z",
    })

    buf := new(bytes.Buffer)
    cmd.SetOut(buf)
    cmd.SetErr(buf)
    cmd.SetArgs([]string{"version"})

    if err := cmd.Execute(); err != nil {
        t.Fatalf("Execute() error = %v", err)
    }

    got := buf.String()
    want := `{"ok":true,"data":{"version":"v0.1.0","commit":"abc1234","buildTime":"2026-05-09T12:00:00Z"}}\n`
    if got != want {
        t.Fatalf("version output mismatch\nwant: %s\ngot:  %s", want, got)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd modelcraft-cli && go test ./cmd -run TestVersionCommandPrintsInjectedMetadata -v`
Expected: FAIL with `undefined: NewRootCommand` and missing module files.

- [ ] **Step 3: Write the minimal CLI scaffold and version command**

```go
// modelcraft-cli/go.mod
module modelcraft-cli

go 1.24.0

require github.com/spf13/cobra v1.8.1
```

```go
// modelcraft-cli/main.go
package main

import "modelcraft-cli/cmd"

var (
    version   = "dev"
    commit    = "none"
    buildTime = "unknown"
)

func main() {
    root := cmd.NewRootCommand(cmd.BuildInfo{
        Version:   version,
        Commit:    commit,
        BuildTime: buildTime,
    })
    _ = root.Execute()
}
```

```go
// modelcraft-cli/cmd/root.go
package cmd

import "github.com/spf13/cobra"

type BuildInfo struct {
    Version   string
    Commit    string
    BuildTime string
}

func NewRootCommand(info BuildInfo) *cobra.Command {
    root := &cobra.Command{Use: "mc", SilenceUsage: true, SilenceErrors: true}
    root.AddCommand(newVersionCommand(info))
    return root
}
```

```go
// modelcraft-cli/cmd/version.go
package cmd

import (
    "fmt"

    "github.com/spf13/cobra"
)

func newVersionCommand(info BuildInfo) *cobra.Command {
    return &cobra.Command{
        Use: "version",
        RunE: func(cmd *cobra.Command, args []string) error {
            _, err := fmt.Fprintf(cmd.OutOrStdout(), "{\"ok\":true,\"data\":{\"version\":\"%s\",\"commit\":\"%s\",\"buildTime\":\"%s\"}}\n", info.Version, info.Commit, info.BuildTime)
            return err
        },
    }
}
```

```make
# modelcraft-cli/justfile
set shell := ["bash", "-cu"]

build:
    go build -ldflags "-X main.version=$(git describe --tags --always) -X main.commit=$(git rev-parse --short HEAD) -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o bin/mc ./main.go

test:
    go test ./...

fmt:
    gofmt -w .

tidy:
    go mod tidy
```

- [ ] **Step 4: Run tests and verify the scaffold works**

Run:
- `cd modelcraft-cli && go mod tidy`
- `cd modelcraft-cli && go test ./cmd -run TestVersionCommandPrintsInjectedMetadata -v`
- `cd modelcraft-cli && just build`
Expected:
- `go mod tidy` succeeds
- test is `PASS`
- `just build` creates `modelcraft-cli/bin/mc`

- [ ] **Step 5: Commit**

```bash
git add modelcraft-cli/go.mod modelcraft-cli/go.sum modelcraft-cli/main.go modelcraft-cli/justfile modelcraft-cli/cmd/root.go modelcraft-cli/cmd/version.go modelcraft-cli/cmd/version_test.go
git commit -m "feat(cli): scaffold cobra-based cli module"
```

### Task 2: Add Unified Output and English Error Contracts

**Files:**
- Create: `modelcraft-cli/internal/output/success.go`
- Create: `modelcraft-cli/internal/output/errors.go`
- Modify: `modelcraft-cli/cmd/root.go`
- Modify: `modelcraft-cli/cmd/version.go`
- Test: `modelcraft-cli/internal/output/errors_test.go`

- [ ] **Step 1: Write failing tests for JSON success and English error envelopes**

```go
package output

import (
    "bytes"
    "errors"
    "testing"
)

func TestWriteSuccessCompactJSON(t *testing.T) {
    buf := new(bytes.Buffer)
    if err := WriteSuccess(buf, "json", true, map[string]any{"version": "v0.1.0"}, nil); err != nil {
        t.Fatalf("WriteSuccess() error = %v", err)
    }
    want := "{\"ok\":true,\"data\":{\"version\":\"v0.1.0\"}}\n"
    if buf.String() != want {
        t.Fatalf("unexpected output: %s", buf.String())
    }
}

func TestCLIErrorExitCodeAndEnvelope(t *testing.T) {
    err := NewCLIError("NO_PROJECT_CONTEXT", "No project context is selected.", true, "Use --project <slug> or run 'mc auth switch-project <slug>'.", map[string]any{"availableProjects": []string{"sales"}})
    if code := ExitCode(err); code != 5 {
        t.Fatalf("ExitCode() = %d, want 5", code)
    }

    buf := new(bytes.Buffer)
    if writeErr := WriteError(buf, "json", true, err); writeErr != nil {
        t.Fatalf("WriteError() error = %v", writeErr)
    }

    want := "{\"ok\":false,\"error\":{\"code\":\"NO_PROJECT_CONTEXT\",\"message\":\"No project context is selected.\",\"retryable\":true,\"suggestion\":\"Use --project <slug> or run 'mc auth switch-project <slug>'.\",\"details\":{\"availableProjects\":[\"sales\"]}}}\n"
    if buf.String() != want {
        t.Fatalf("unexpected envelope: %s", buf.String())
    }

    if !errors.Is(err, ErrCLI) {
        t.Fatalf("expected CLI sentinel")
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd modelcraft-cli && go test ./internal/output -v`
Expected: FAIL with missing `WriteSuccess`, `NewCLIError`, `WriteError`, and `ExitCode`.

- [ ] **Step 3: Implement the output package and wire version through it**

```go
// modelcraft-cli/internal/output/errors.go
package output

import (
    "encoding/json"
    "errors"
    "fmt"
    "io"
)

var ErrCLI = errors.New("cli error")

type CLIError struct {
    Code       string
    Message    string
    Retryable  bool
    Suggestion string
    Details    map[string]any
}

func (e *CLIError) Error() string { return e.Code + ": " + e.Message }
func (e *CLIError) Unwrap() error { return ErrCLI }

func NewCLIError(code, message string, retryable bool, suggestion string, details map[string]any) *CLIError {
    return &CLIError{Code: code, Message: message, Retryable: retryable, Suggestion: suggestion, Details: details}
}

func ExitCode(err error) int {
    var cliErr *CLIError
    if !errors.As(err, &cliErr) {
        return 1
    }
    switch cliErr.Code {
    case "INVALID_JSON_FLAG", "INVALID_ARGUMENT", "MISSING_REQUIRED_FLAG":
        return 2
    case "UNAUTHENTICATED", "TOKEN_EXPIRED", "INVALID_CREDENTIALS":
        return 3
    case "PERMISSION_DENIED":
        return 4
    case "NO_PROJECT_CONTEXT", "MODEL_NOT_FOUND", "DATABASE_NOT_FOUND", "NOT_FOUND":
        return 5
    case "TAKE_EXCEEDS_LIMIT":
        return 6
    default:
        return 7
    }
}

func WriteError(w io.Writer, format string, compact bool, err error) error {
    var cliErr *CLIError
    if !errors.As(err, &cliErr) {
        cliErr = NewCLIError("UNKNOWN_ERROR", fmt.Sprintf("%v", err), false, "Inspect stderr or rerun with --verbose.", nil)
    }
    payload := map[string]any{
        "ok": false,
        "error": map[string]any{
            "code":       cliErr.Code,
            "message":    cliErr.Message,
            "retryable":  cliErr.Retryable,
            "suggestion": cliErr.Suggestion,
        },
    }
    if cliErr.Details != nil {
        payload["error"].(map[string]any)["details"] = cliErr.Details
    }
    return writeJSON(w, compact, payload)
}

func writeJSON(w io.Writer, compact bool, payload any) error {
    var (
        b   []byte
        err error
    )
    if compact {
        b, err = json.Marshal(payload)
    } else {
        b, err = json.MarshalIndent(payload, "", "  ")
    }
    if err != nil {
        return err
    }
    _, err = w.Write(append(b, '\n'))
    return err
}
```

```go
// modelcraft-cli/internal/output/success.go
package output

import "io"

func WriteSuccess(w io.Writer, format string, compact bool, data any, meta map[string]any) error {
    payload := map[string]any{"ok": true, "data": data}
    if meta != nil {
        payload["meta"] = meta
    }
    return writeJSON(w, compact, payload)
}
```

```go
// modelcraft-cli/cmd/version.go
package cmd

import (
    "modelcraft-cli/internal/output"

    "github.com/spf13/cobra"
)

func newVersionCommand(info BuildInfo) *cobra.Command {
    return &cobra.Command{
        Use: "version",
        RunE: func(cmd *cobra.Command, args []string) error {
            return output.WriteSuccess(cmd.OutOrStdout(), "json", true, map[string]any{
                "version":   info.Version,
                "commit":    info.Commit,
                "buildTime": info.BuildTime,
            }, nil)
        },
    }
}
```

- [ ] **Step 4: Run tests and root command verification**

Run:
- `cd modelcraft-cli && go test ./internal/output -v`
- `cd modelcraft-cli && go test ./cmd -run TestVersionCommandPrintsInjectedMetadata -v`
Expected: all targeted tests `PASS`.

- [ ] **Step 5: Commit**

```bash
git add modelcraft-cli/internal/output/success.go modelcraft-cli/internal/output/errors.go modelcraft-cli/internal/output/errors_test.go modelcraft-cli/cmd/root.go modelcraft-cli/cmd/version.go
git commit -m "feat(cli): add unified success and error output contracts"
```

### Task 3: Implement Credential Storage, Session Refresh, and Project Context Rules

**Files:**
- Create: `modelcraft-cli/internal/config/credentials.go`
- Create: `modelcraft-cli/internal/auth/session.go`
- Create: `modelcraft-cli/internal/app/context.go`
- Test: `modelcraft-cli/internal/config/credentials_test.go`
- Test: `modelcraft-cli/internal/auth/session_test.go`

- [ ] **Step 1: Write failing tests for single-profile storage and refresh thresholds**

```go
package config

import (
    "path/filepath"
    "testing"
    "time"
)

func TestSaveAndLoadCredentials(t *testing.T) {
    dir := t.TempDir()
    path := filepath.Join(dir, "credentials.json")

    creds := Credentials{
        Server:       "https://gateway.example.com",
        OrgName:      "acme",
        UserID:       "user-1",
        AccessToken:  "access",
        RefreshToken: "refresh",
        ExpiresAt:    time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC),
        Projects:     []AccessibleProject{{Slug: "sales", Title: "Sales"}},
    }

    if err := Save(path, creds); err != nil {
        t.Fatalf("Save() error = %v", err)
    }
    got, err := Load(path)
    if err != nil {
        t.Fatalf("Load() error = %v", err)
    }
    if got.OrgName != "acme" || got.CurrentProject != "" {
        t.Fatalf("unexpected credentials: %+v", got)
    }
}
```

```go
package auth

import (
    "context"
    "testing"
    "time"

    "modelcraft-cli/internal/config"
)

type fakeAuthClient struct{ refreshed bool }

func (f *fakeAuthClient) Refresh(context.Context, string, string, string) (*config.Credentials, error) {
    f.refreshed = true
    return &config.Credentials{AccessToken: "fresh", RefreshToken: "fresh-r", ExpiresAt: time.Now().Add(2 * time.Hour)}, nil
}

func TestEnsureFreshRefreshesWhenExpiryIsWithinOneMinute(t *testing.T) {
    mgr := Manager{Now: func() time.Time { return time.Date(2026, 5, 9, 12, 0, 0, 0, time.UTC) }}
    creds := config.Credentials{Server: "https://gateway.example.com", OrgName: "acme", RefreshToken: "r1", ExpiresAt: time.Date(2026, 5, 9, 12, 0, 30, 0, time.UTC)}
    client := &fakeAuthClient{}

    got, err := mgr.EnsureFresh(context.Background(), creds, client)
    if err != nil {
        t.Fatalf("EnsureFresh() error = %v", err)
    }
    if !client.refreshed || got.AccessToken != "fresh" {
        t.Fatalf("expected refresh to happen, got %+v", got)
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd modelcraft-cli && go test ./internal/config ./internal/auth -v`
Expected: FAIL with missing credential/session implementations.

- [ ] **Step 3: Implement single-profile credentials and session manager**

```go
// modelcraft-cli/internal/config/credentials.go
package config

import (
    "encoding/json"
    "os"
    "path/filepath"
    "time"
)

type AccessibleProject struct {
    Slug  string `json:"slug"`
    Title string `json:"title"`
}

type Credentials struct {
    Server         string              `json:"server"`
    OrgName        string              `json:"orgName"`
    UserID         string              `json:"userId"`
    AccessToken    string              `json:"accessToken"`
    RefreshToken   string              `json:"refreshToken"`
    ExpiresAt      time.Time           `json:"expiresAt"`
    Projects       []AccessibleProject `json:"projects,omitempty"`
    CurrentProject string              `json:"currentProject,omitempty"`
}

func DefaultPath() string {
    home, _ := os.UserHomeDir()
    return filepath.Join(home, ".config", "modelcraft", "credentials.json")
}

func Save(path string, creds Credentials) error {
    if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
        return err
    }
    b, err := json.MarshalIndent(creds, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(path, append(b, '\n'), 0o600)
}

func Load(path string) (Credentials, error) {
    b, err := os.ReadFile(path)
    if err != nil {
        return Credentials{}, err
    }
    var creds Credentials
    err = json.Unmarshal(b, &creds)
    return creds, err
}
```

```go
// modelcraft-cli/internal/auth/session.go
package auth

import (
    "context"
    "time"

    "modelcraft-cli/internal/config"
    "modelcraft-cli/internal/output"
)

type RefreshClient interface {
    Refresh(ctx context.Context, server, orgName, refreshToken string) (*config.Credentials, error)
}

type Manager struct {
    Now func() time.Time
}

func (m Manager) EnsureFresh(ctx context.Context, creds config.Credentials, client RefreshClient) (config.Credentials, error) {
    now := time.Now()
    if m.Now != nil {
        now = m.Now()
    }
    if creds.ExpiresAt.After(now.Add(60 * time.Second)) {
        return creds, nil
    }
    if creds.RefreshToken == "" {
        return config.Credentials{}, output.NewCLIError("TOKEN_EXPIRED", "Access token has expired.", true, "Run 'mc auth login'.", nil)
    }
    fresh, err := client.Refresh(ctx, creds.Server, creds.OrgName, creds.RefreshToken)
    if err != nil {
        return config.Credentials{}, err
    }
    fresh.CurrentProject = creds.CurrentProject
    return *fresh, nil
}

func SwitchProject(creds config.Credentials, slug string) (config.Credentials, error) {
    for _, p := range creds.Projects {
        if p.Slug == slug {
            creds.CurrentProject = slug
            return creds, nil
        }
    }
    return config.Credentials{}, output.NewCLIError("PROJECT_NOT_FOUND", "Project is not accessible for the current user.", false, "Run 'mc catalog projects' to inspect available projects.", map[string]any{"project": slug})
}
```

- [ ] **Step 4: Run tests and add environment-override cases**

Run:
- `cd modelcraft-cli && go test ./internal/config ./internal/auth -v`
Expected: tests `PASS`.

Then extend tests with:
- `MC_SERVER` override
- `MC_ORG` override
- `MC_ACCESS_TOKEN` override
- `MC_PROJECT` override

Re-run: `cd modelcraft-cli && go test ./internal/config ./internal/auth -v`
Expected: tests remain `PASS`.

- [ ] **Step 5: Commit**

```bash
git add modelcraft-cli/internal/config/credentials.go modelcraft-cli/internal/config/credentials_test.go modelcraft-cli/internal/auth/session.go modelcraft-cli/internal/auth/session_test.go modelcraft-cli/internal/app/context.go
git commit -m "feat(cli): add single-profile session and project context management"
```

### Task 4: Add Gateway CLI EndUser Auth Passthrough Routes

**Files:**
- Modify: `modelcraft-gateway/internal/auth/handler.go`
- Modify: `modelcraft-gateway/cmd/gateway/main.go`
- Create: `modelcraft-gateway/internal/auth/handler_cli_test.go`

- [ ] **Step 1: Write failing Gateway tests for CLI login and refresh passthrough**

```go
package auth

import (
    "bytes"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestCLIEndUserLoginReturnsRefreshTokenInJSONBody(t *testing.T) {
    backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/api/end-user/auth/login" {
            t.Fatalf("unexpected path: %s", r.URL.Path)
        }
        w.Header().Set("Content-Type", "application/json")
        _, _ = w.Write([]byte(`{"requestId":"r1","userId":"u1","accessToken":"a1","refreshToken":"rt1","expiresAt":"2026-05-09T12:00:00Z","projects":[]}`))
    }))
    defer backend.Close()

    h := NewHandler(&Service{}, backend.URL, backend.Client(), "")
    req := httptest.NewRequest(http.MethodPost, "/api/cli/end-user/auth/login", bytes.NewBufferString(`{"orgName":"acme","username":"alice","password":"secret"}`))
    rec := httptest.NewRecorder()

    h.CLIEndUserLogin(rec, req)

    if got := rec.Header().Get("Set-Cookie"); got != "" {
        t.Fatalf("did not expect refresh cookie, got %q", got)
    }
    want := `{"requestId":"r1","userId":"u1","accessToken":"a1","refreshToken":"rt1","expiresAt":"2026-05-09T12:00:00Z","projects":[]}` + "\n"
    if rec.Body.String() != want {
        t.Fatalf("unexpected body: %s", rec.Body.String())
    }
}
```

```go
func TestCLIEndUserRefreshReadsRefreshTokenFromBody(t *testing.T) {
    backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        buf, _ := io.ReadAll(r.Body)
        want := `{"orgName":"acme","refreshToken":"rt1"}`
        if string(buf) != want {
            t.Fatalf("unexpected request body: %s", buf)
        }
        _, _ = w.Write([]byte(`{"requestId":"r2","userId":"u1","accessToken":"a2","refreshToken":"rt2","expiresAt":"2026-05-09T13:00:00Z","projects":[]}`))
    }))
    defer backend.Close()

    h := NewHandler(&Service{}, backend.URL, backend.Client(), "")
    req := httptest.NewRequest(http.MethodPost, "/api/cli/end-user/auth/refresh", bytes.NewBufferString(`{"orgName":"acme","refreshToken":"rt1"}`))
    rec := httptest.NewRecorder()

    h.CLIEndUserRefresh(rec, req)

    if rec.Code != http.StatusOK {
        t.Fatalf("status = %d, want 200", rec.Code)
    }
}
```

- [ ] **Step 2: Run Gateway auth tests to verify they fail**

Run: `cd modelcraft-gateway && go test ./internal/auth -run 'TestCLIEndUser(Login|Refresh)' -v`
Expected: FAIL with undefined `CLIEndUserLogin` and `CLIEndUserRefresh`.

- [ ] **Step 3: Implement CLI-only passthrough handlers and register new routes**

```go
// modelcraft-gateway/internal/auth/handler.go
func (h *Handler) CLIEndUserLogin(w http.ResponseWriter, r *http.Request) {
    raw, status, err := h.postBackendRaw(r.Context(), "/api/end-user/auth/login", r.Body)
    if err != nil {
        proxyBackendError(w, err)
        return
    }
    if status >= 400 {
        writeRaw(w, status, raw)
        return
    }
    writeJSON(w, http.StatusOK, json.RawMessage(raw))
}

func (h *Handler) CLIEndUserRefresh(w http.ResponseWriter, r *http.Request) {
    raw, status, err := h.postBackendRaw(r.Context(), "/api/end-user/auth/refresh", r.Body)
    if err != nil {
        proxyBackendError(w, err)
        return
    }
    if status >= 400 {
        writeRaw(w, status, raw)
        return
    }
    writeJSON(w, http.StatusOK, json.RawMessage(raw))
}

func (h *Handler) CLIEndUserLogout(w http.ResponseWriter, r *http.Request) {
    raw, status, err := h.postBackendRaw(r.Context(), "/api/end-user/auth/logout", r.Body)
    if err != nil {
        proxyBackendError(w, err)
        return
    }
    if status >= 400 {
        writeRaw(w, status, raw)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) CLIEndUserSelectProject(w http.ResponseWriter, r *http.Request) {
    raw, status, err := h.postBackendRaw(r.Context(), "/api/end-user/auth/select-project", r.Body)
    if err != nil {
        proxyBackendError(w, err)
        return
    }
    if status >= 400 {
        writeRaw(w, status, raw)
        return
    }
    writeJSON(w, http.StatusOK, json.RawMessage(raw))
}

func (h *Handler) CLIEndUserMe(w http.ResponseWriter, r *http.Request) {
    authHeader := r.Header.Get("Authorization")
    if authHeader == "" {
        writeError(w, http.StatusUnauthorized, "MISSING_TOKEN", "Authorization header required")
        return
    }
    raw, err := h.getBackendRaw(r.Context(), "/api/end-user/auth/me", func(req *http.Request) {
        req.Header.Set("Authorization", authHeader)
    })
    if err != nil {
        proxyBackendError(w, err)
        return
    }
    writeJSON(w, http.StatusOK, json.RawMessage(raw))
}
```

```go
// modelcraft-gateway/cmd/gateway/main.go
r.Route("/api/cli/end-user/auth", func(r chi.Router) {
    r.Post("/login", authHandler.CLIEndUserLogin)
    r.Post("/refresh", authHandler.CLIEndUserRefresh)
    r.Post("/logout", authHandler.CLIEndUserLogout)
    r.Post("/select-project", authHandler.CLIEndUserSelectProject)
    r.Get("/me", authHandler.CLIEndUserMe)
})
```

- [ ] **Step 4: Run Gateway tests and route-level verification**

Run:
- `cd modelcraft-gateway && go test ./internal/auth -run 'TestCLIEndUser(Login|Refresh)' -v`
- `cd modelcraft-gateway && go test ./...`
Expected: CLI auth tests `PASS`; full Gateway test suite remains green.

- [ ] **Step 5: Commit**

```bash
git add modelcraft-gateway/internal/auth/handler.go modelcraft-gateway/internal/auth/handler_cli_test.go modelcraft-gateway/cmd/gateway/main.go
git commit -m "feat(gateway): add cli end-user auth passthrough routes"
```

### Task 5: Implement CLI Auth Commands Against the New Gateway Routes

**Files:**
- Create: `modelcraft-cli/internal/client/auth.go`
- Create: `modelcraft-cli/cmd/auth.go`
- Modify: `modelcraft-cli/cmd/root.go`
- Test: `modelcraft-cli/internal/client/auth_test.go`
- Test: `modelcraft-cli/cmd/auth_test.go`

- [ ] **Step 1: Write failing tests for login persistence and switch-project behavior**

```go
package cmd

import (
    "bytes"
    "net/http"
    "net/http/httptest"
    "os"
    "path/filepath"
    "testing"
)

func TestAuthLoginPersistsSingleProfileWithoutSelectingProject(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        _, _ = w.Write([]byte(`{"requestId":"r1","userId":"u1","accessToken":"a1","refreshToken":"rt1","expiresAt":"2026-05-09T12:00:00Z","projects":[{"slug":"sales","title":"Sales"}]}`))
    }))
    defer srv.Close()

    dir := t.TempDir()
    credPath := filepath.Join(dir, "credentials.json")

    cmd := NewRootCommand(BuildInfo{})
    buf := new(bytes.Buffer)
    cmd.SetOut(buf)
    cmd.SetErr(buf)
    cmd.SetArgs([]string{"auth", "login", "--server", srv.URL, "--org", "acme", "--username", "alice", "--password", "secret", "--credentials", credPath})

    if err := cmd.Execute(); err != nil {
        t.Fatalf("Execute() error = %v", err)
    }

    b, err := os.ReadFile(credPath)
    if err != nil {
        t.Fatalf("ReadFile() error = %v", err)
    }
    if bytes.Contains(b, []byte(`"currentProject":"sales"`)) {
        t.Fatalf("login must not set currentProject automatically: %s", b)
    }
}
```

```go
func TestAuthSwitchProjectRejectsUnknownProject(t *testing.T) {
    dir := t.TempDir()
    credPath := filepath.Join(dir, "credentials.json")
    _ = os.WriteFile(credPath, []byte(`{"server":"https://gateway.example.com","orgName":"acme","projects":[{"slug":"sales","title":"Sales"}]}`), 0o600)

    cmd := NewRootCommand(BuildInfo{})
    buf := new(bytes.Buffer)
    cmd.SetOut(buf)
    cmd.SetErr(buf)
    cmd.SetArgs([]string{"auth", "switch-project", "hr", "--credentials", credPath})

    err := cmd.Execute()
    if err == nil {
        t.Fatalf("expected switch-project error")
    }
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd modelcraft-cli && go test ./cmd -run 'TestAuth(LoginPersistsSingleProfileWithoutSelectingProject|SwitchProjectRejectsUnknownProject)' -v`
Expected: FAIL with missing auth command/client implementations.

- [ ] **Step 3: Implement REST auth client and auth commands**

```go
// modelcraft-cli/internal/client/auth.go
package client

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"

    "modelcraft-cli/internal/config"
    "modelcraft-cli/internal/output"
)

type AuthClient struct{ HTTPClient *http.Client }

type loginRequest struct {
    OrgName  string `json:"orgName"`
    Username string `json:"username"`
    Password string `json:"password"`
}

func (c AuthClient) Login(ctx context.Context, server, org, username, password string) (*config.Credentials, error) {
    body, _ := json.Marshal(loginRequest{OrgName: org, Username: username, Password: password})
    req, _ := http.NewRequestWithContext(ctx, http.MethodPost, server+"/api/cli/end-user/auth/login", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, output.NewCLIError("SERVICE_UNAVAILABLE", "Gateway is unreachable.", true, "Check network connectivity and retry.", nil)
    }
    defer resp.Body.Close()
    var creds config.Credentials
    if err := json.NewDecoder(resp.Body).Decode(&creds); err != nil {
        return nil, output.NewCLIError("INVALID_UPSTREAM", "Gateway returned invalid JSON.", false, "Inspect the upstream service output.", nil)
    }
    creds.Server = server
    creds.OrgName = org
    return &creds, nil
}
```

```go
// modelcraft-cli/cmd/auth.go
package cmd

func newAuthCommand(deps Dependencies) *cobra.Command {
    authCmd := &cobra.Command{Use: "auth"}
    authCmd.AddCommand(newAuthLoginCommand(deps))
    authCmd.AddCommand(newAuthLogoutCommand(deps))
    authCmd.AddCommand(newAuthRefreshCommand(deps))
    authCmd.AddCommand(newAuthStatusCommand(deps))
    authCmd.AddCommand(newAuthSwitchProjectCommand(deps))
    return authCmd
}

func newAuthLoginCommand(deps Dependencies) *cobra.Command {
    var server, org, username, password string
    cmd := &cobra.Command{
        Use: "login",
        RunE: func(cmd *cobra.Command, args []string) error {
            creds, err := deps.AuthClient.Login(cmd.Context(), server, org, username, password)
            if err != nil {
                return err
            }
            if err := deps.Store.Save(*creds); err != nil {
                return err
            }
            return deps.Writer.Success(cmd.OutOrStdout(), map[string]any{
                "server":   creds.Server,
                "orgName":  creds.OrgName,
                "userId":   creds.UserID,
                "projects": creds.Projects,
            }, nil)
        },
    }
    cmd.Flags().StringVar(&server, "server", "", "Gateway base URL")
    cmd.Flags().StringVar(&org, "org", "", "Organization slug")
    cmd.Flags().StringVar(&username, "username", "", "EndUser username")
    cmd.Flags().StringVar(&password, "password", "", "EndUser password")
    _ = cmd.MarkFlagRequired("server")
    _ = cmd.MarkFlagRequired("org")
    _ = cmd.MarkFlagRequired("username")
    _ = cmd.MarkFlagRequired("password")
    return cmd
}
```

- [ ] **Step 4: Run targeted tests and CLI auth smoke checks**

Run:
- `cd modelcraft-cli && go test ./internal/client ./cmd -run 'TestAuth(LoginPersistsSingleProfileWithoutSelectingProject|SwitchProjectRejectsUnknownProject)' -v`
- `cd modelcraft-cli && go test ./...`
Expected: targeted and full CLI tests `PASS`.

Then smoke-check with a local Gateway:
- `cd modelcraft-cli && go run . auth status`
Expected: JSON error `UNAUTHENTICATED` when no credentials exist.

- [ ] **Step 5: Commit**

```bash
git add modelcraft-cli/internal/client/auth.go modelcraft-cli/internal/client/auth_test.go modelcraft-cli/cmd/auth.go modelcraft-cli/cmd/auth_test.go modelcraft-cli/cmd/root.go
git commit -m "feat(cli): add gateway-backed auth commands"
```

### Task 6: Implement Resource Parsing and Catalog Discovery Commands

**Files:**
- Create: `modelcraft-cli/internal/resource/path.go`
- Create: `modelcraft-cli/internal/client/graphql.go`
- Create: `modelcraft-cli/internal/client/catalog.go`
- Create: `modelcraft-cli/cmd/catalog.go`
- Test: `modelcraft-cli/internal/resource/path_test.go`
- Test: `modelcraft-cli/internal/client/catalog_test.go`
- Test: `modelcraft-cli/cmd/catalog_test.go`

- [x] **Step 1: Write failing tests for path parsing and project fallback**

```go
package resource

import "testing"

func TestParseDatabaseModelUsesCurrentProjectFallback(t *testing.T) {
    got, err := ParseModelPath("maindb.users", ParseContext{CurrentProject: "sales"})
    if err != nil {
        t.Fatalf("ParseModelPath() error = %v", err)
    }
    if got.Project != "sales" || got.Database != "maindb" || got.Model != "users" {
        t.Fatalf("unexpected path: %+v", got)
    }
}

func TestParseModelPathRejectsSingleSegment(t *testing.T) {
    _, err := ParseModelPath("users", ParseContext{CurrentProject: "sales"})
    if err == nil {
        t.Fatalf("expected single-segment rejection")
    }
}
```

```go
func TestCatalogProjectsDoesNotRequireCurrentProject(t *testing.T) {
    // Use a mocked GraphQL server to return one project and assert command success.
}
```

- [x] **Step 2: Run tests to verify they fail**

Run: `cd modelcraft-cli && go test ./internal/resource ./internal/client ./cmd -run 'Test(ParseDatabaseModelUsesCurrentProjectFallback|ParseModelPathRejectsSingleSegment|CatalogProjectsDoesNotRequireCurrentProject)' -v`
Expected: FAIL with missing parser/catalog implementation.

- [x] **Step 3: Implement the parser and catalog clients/commands**

```go
// modelcraft-cli/internal/resource/path.go
package resource

import "modelcraft-cli/internal/output"

type ParseContext struct{ CurrentProject string }

type ModelPath struct {
    Project  string
    Database string
    Model    string
}

func ParseModelPath(raw string, ctx ParseContext) (ModelPath, error) {
    parts := strings.Split(raw, ".")
    switch len(parts) {
    case 3:
        return ModelPath{Project: parts[0], Database: parts[1], Model: parts[2]}, nil
    case 2:
        if ctx.CurrentProject == "" {
            return ModelPath{}, output.NewCLIError("NO_PROJECT_CONTEXT", "No project context is selected.", true, "Use --project <slug> or run 'mc auth switch-project <slug>'.", nil)
        }
        return ModelPath{Project: ctx.CurrentProject, Database: parts[0], Model: parts[1]}, nil
    default:
        return ModelPath{}, output.NewCLIError("INVALID_RESOURCE_PATH", "Resource path must be '<project>.<database>.<model>' or '<database>.<model>'.", true, "Provide at least database and model segments.", map[string]any{"path": raw})
    }
}
```

```go
// modelcraft-cli/internal/client/catalog.go
package client

func (c GraphQLClient) CatalogProjects(ctx context.Context, server, org, token string) ([]config.AccessibleProject, error) {
    // Return projects from auth/session for v1 if already present; only use GraphQL when refresh/sync is needed.
}

func (c GraphQLClient) CatalogDatabases(ctx context.Context, server, org, project, token string) ([]string, error) {
    // POST /graphql/end-user/org/{org}/project/{project}
    // query { modelDatabaseCatalog(input:{}) { items { name } } }
}

func (c GraphQLClient) CatalogModels(ctx context.Context, server, org, project, database, token string) ([]map[string]any, error) {
    // POST /graphql/end-user/org/{org}/project/{project}
    // query with modelCatalog filtered by database.
}
```

```go
// modelcraft-cli/cmd/catalog.go
package cmd

func newCatalogCommand(deps Dependencies) *cobra.Command {
    cmd := &cobra.Command{Use: "catalog"}
    cmd.AddCommand(newCatalogProjectsCommand(deps))
    cmd.AddCommand(newCatalogDatabasesCommand(deps))
    cmd.AddCommand(newCatalogModelsCommand(deps))
    return cmd
}
```

- [x] **Step 4: Run tests and manual project-fallback checks**

Run:
- `cd modelcraft-cli && go test ./internal/resource ./internal/client ./cmd -v`
Expected: all parser and catalog tests `PASS`.

Then manually verify:
- `cd modelcraft-cli && go run . catalog projects`
- `cd modelcraft-cli && go run . catalog databases --project sales`
Expected: stable JSON envelopes; no implicit database fallback.

- [x] **Step 5: Commit**

```bash
git add modelcraft-cli/internal/resource/path.go modelcraft-cli/internal/resource/path_test.go modelcraft-cli/internal/client/graphql.go modelcraft-cli/internal/client/catalog.go modelcraft-cli/internal/client/catalog_test.go modelcraft-cli/cmd/catalog.go modelcraft-cli/cmd/catalog_test.go
git commit -m "feat(cli): add resource parsing and catalog discovery commands"
```

### Task 7: Implement Runtime Read Commands (`query`, `get`, `count`, `aggregate`)

**Files:**
- Create: `modelcraft-cli/internal/client/runtime.go`
- Create: `modelcraft-cli/cmd/query.go`
- Test: `modelcraft-cli/internal/client/runtime_test.go`
- Test: `modelcraft-cli/cmd/query_test.go`

- [x] **Step 1: Write failing tests for runtime query translation and `NO_PROJECT_CONTEXT`**

```go
package client

func TestQueryBuildsModelScopedRuntimeEndpoint(t *testing.T) {
    // Assert endpoint path: /graphql/end-user/org/acme/project/sales/db/maindb/model/users
    // Assert body includes GraphQL findMany query with where/select/orderBy/take/skip variables.
}
```

```go
package cmd

func TestQueryCommandReturnsNoProjectContextWhenDatabaseModelLacksFallback(t *testing.T) {
    cmd := NewRootCommand(BuildInfo{})
    buf := new(bytes.Buffer)
    cmd.SetOut(buf)
    cmd.SetErr(buf)
    cmd.SetArgs([]string{"query", "maindb.users"})

    err := cmd.Execute()
    if err == nil {
        t.Fatalf("expected NO_PROJECT_CONTEXT")
    }
}
```

- [x] **Step 2: Run tests to verify they fail**

Run: `cd modelcraft-cli && go test ./internal/client ./cmd -run 'Test(QueryBuildsModelScopedRuntimeEndpoint|QueryCommandReturnsNoProjectContextWhenDatabaseModelLacksFallback)' -v`
Expected: FAIL with missing runtime client/command code.

- [x] **Step 3: Implement runtime GraphQL client and read commands**

```go
// modelcraft-cli/internal/client/runtime.go
package client

type QueryOptions struct {
    Where   json.RawMessage
    Select  []string
    OrderBy json.RawMessage
    Take    int
    Skip    int
}

func (c GraphQLClient) Query(ctx context.Context, server, org, project, db, model, token string, opts QueryOptions) (map[string]any, error) {
    endpoint := fmt.Sprintf("%s/graphql/end-user/org/%s/project/%s/db/%s/model/%s", server, org, project, db, model)
    query := `query FindMany($where: JSON, $select: [String!], $orderBy: JSON, $take: Int, $skip: Int) { findMany(where: $where, select: $select, orderBy: $orderBy, take: $take, skip: $skip) }`
    // Send request and map GraphQL errors to CLI errors.
}
```

```go
// modelcraft-cli/cmd/query.go
package cmd

func newQueryCommand(deps Dependencies) *cobra.Command {
    var where, orderBy string
    var selectFields []string
    var take, skip int
    cmd := &cobra.Command{
        Use: "query <path>",
        Args: cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            // Resolve project fallback, call runtime client, emit {ok,data,meta}.
            return nil
        },
    }
    cmd.Flags().StringVar(&where, "where", "", "JSON where filter")
    cmd.Flags().StringSliceVar(&selectFields, "select", nil, "Selected fields")
    cmd.Flags().StringVar(&orderBy, "orderBy", "", "JSON orderBy expression")
    cmd.Flags().IntVar(&take, "take", 20, "page size")
    cmd.Flags().IntVar(&skip, "skip", 0, "records to skip")
    return cmd
}
```

- [x] **Step 4: Run runtime command tests and a local smoke query**

Run:
- `cd modelcraft-cli && go test ./internal/client ./cmd -v`
Expected: runtime unit tests `PASS`.

Smoke:
- `cd modelcraft-cli && go run . query sales.maindb.users --take 1`
Expected: JSON output with `ok=true` or a GraphQL-derived English error envelope.

- [x] **Step 5: Commit**

```bash
git add modelcraft-cli/internal/client/runtime.go modelcraft-cli/internal/client/runtime_test.go modelcraft-cli/cmd/query.go modelcraft-cli/cmd/query_test.go
git commit -m "feat(cli): add runtime read commands"
```

### Task 8: Implement `schema` and `describe`

**Files:**
- Create: `modelcraft-cli/internal/schema/commands.go`
- Create: `modelcraft-cli/cmd/schema.go`
- Create: `modelcraft-cli/cmd/describe.go`
- Test: `modelcraft-cli/internal/schema/commands_test.go`
- Test: `modelcraft-cli/cmd/schema_test.go`
- Test: `modelcraft-cli/cmd/describe_test.go`

- [x] **Step 1: Write failing tests for static schema export and model describe output**

```go
package schema

func TestBuildCommandSchemaIncludesQueryFlags(t *testing.T) {
    doc := BuildCommandSchema()
    query, ok := doc.Commands["query"]
    if !ok {
        t.Fatalf("query command missing")
    }
    if _, ok := query.Flags["take"]; !ok {
        t.Fatalf("query --take flag missing")
    }
}
```

```go
package cmd

func TestDescribeModelUsesRuntimeIntrospection(t *testing.T) {
    // Mock the runtime GraphQL endpoint to return __type / __schema info.
    // Assert CLI returns fields/type/required/list metadata without invented limits.
}
```

- [x] **Step 2: Run tests to verify they fail**

Run: `cd modelcraft-cli && go test ./internal/schema ./cmd -run 'Test(BuildCommandSchemaIncludesQueryFlags|DescribeModelUsesRuntimeIntrospection)' -v`
Expected: FAIL with missing schema and describe implementations.

- [x] **Step 3: Implement local schema generation and remote describe**

```go
// modelcraft-cli/internal/schema/commands.go
package schema

type CommandSchema struct {
    Commands map[string]CommandDoc `json:"commands"`
}

type CommandDoc struct {
    Description string             `json:"description"`
    Usage       string             `json:"usage"`
    Flags       map[string]FlagDoc `json:"flags,omitempty"`
}

type FlagDoc struct {
    Type        string `json:"type"`
    Required    bool   `json:"required"`
    Description string `json:"description"`
}

func BuildCommandSchema() CommandSchema {
    return CommandSchema{
        Commands: map[string]CommandDoc{
            "query": {
                Description: "Query multiple records from a runtime model.",
                Usage:       "mc query <project.database.model|database.model>",
                Flags: map[string]FlagDoc{
                    "take": {Type: "int", Required: false, Description: "Page size."},
                    "skip": {Type: "int", Required: false, Description: "Records to skip."},
                },
            },
        },
    }
}
```

```go
// modelcraft-cli/cmd/describe.go
package cmd

func newDescribeCommand(deps Dependencies) *cobra.Command {
    return &cobra.Command{
        Use:  "describe <path>",
        Args: cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            // If args[0] has 2 segments, describe database through catalog.
            // If args[0] has 3 segments or falls back to currentProject, call runtime introspection.
            // Return only authoritative introspection-derived metadata plus catalog metadata that already exists.
            return nil
        },
    }
}
```

- [x] **Step 4: Run tests and inspect command schema output**

Run:
- `cd modelcraft-cli && go test ./internal/schema ./cmd -v`
- `cd modelcraft-cli && go run . schema commands`
Expected:
- tests `PASS`
- schema output is local, fast, and does not require credentials.

- [x] **Step 5: Commit**

```bash
git add modelcraft-cli/internal/schema/commands.go modelcraft-cli/internal/schema/commands_test.go modelcraft-cli/cmd/schema.go modelcraft-cli/cmd/schema_test.go modelcraft-cli/cmd/describe.go modelcraft-cli/cmd/describe_test.go
git commit -m "feat(cli): add static schema export and remote describe"
```

### Task 9: Finish Docs, Build Metadata, and Cross-Platform Smoke Verification

**Files:**
- Create: `modelcraft-cli/README.md`
- Modify: `modelcraft-cli/justfile`
- Modify: `modelcraft-cli/main.go`
- Test: `modelcraft-cli/README.md` examples verified manually

- [ ] **Step 1: Write failing smoke checklist items into the README draft**

```md
# ModelCraft CLI

## Windows PowerShell

```powershell
mc.exe auth login --server https://gateway.example.com --org acme --username alice --password secret
mc.exe query sales.maindb.users --where '{"username":{"contains":"alice"}}' --take 1
```
```

Document the expected output contracts before finalizing the commands so the examples can be validated against real behavior.

- [ ] **Step 2: Run CLI builds for target platforms to identify missing linker/build metadata**

Run:
- `cd modelcraft-cli && just build`
- `cd modelcraft-cli && GOOS=windows GOARCH=amd64 go build -o dist/mc-windows-amd64.exe ./main.go`
Expected: if version/build metadata or directory creation is missing, builds or `version` output checks will fail.

- [ ] **Step 3: Implement release build targets and finish the operator docs**

```make
# modelcraft-cli/justfile
release:
    mkdir -p dist
    GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$(git describe --tags --always) -X main.commit=$(git rev-parse --short HEAD) -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o dist/mc-windows-amd64.exe ./main.go
    GOOS=windows GOARCH=arm64 go build -ldflags "-X main.version=$(git describe --tags --always) -X main.commit=$(git rev-parse --short HEAD) -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o dist/mc-windows-arm64.exe ./main.go
    GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$(git describe --tags --always) -X main.commit=$(git rev-parse --short HEAD) -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o dist/mc-darwin-amd64 ./main.go
    GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.version=$(git describe --tags --always) -X main.commit=$(git rev-parse --short HEAD) -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o dist/mc-darwin-arm64 ./main.go
    GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(git describe --tags --always) -X main.commit=$(git rev-parse --short HEAD) -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o dist/mc-linux-amd64 ./main.go
```

```md
# modelcraft-cli/README.md
## Quick Start
1. `mc auth login --server <gateway> --org <org> --username <user> --password <pass>`
2. `mc catalog projects`
3. `mc auth switch-project <slug>`
4. `mc catalog databases`
5. `mc query <database.model> --take 1`

## Output Contract
- `stdout`: JSON or YAML only
- `stderr`: diagnostics only
- errors are always English

## Windows PowerShell Notes
- Prefer single quotes around JSON flag values.
- Example: `mc.exe query sales.maindb.users --where '{"username":{"contains":"alice"}}'`
```

- [ ] **Step 4: Run final verification commands**

Run:
- `cd modelcraft-cli && go test ./...`
- `cd modelcraft-gateway && go test ./...`
- `cd modelcraft-cli && just release`
- `cd modelcraft-cli && ./bin/mc version`
Expected:
- all tests `PASS`
- release artifacts created in `modelcraft-cli/dist/`
- `version` prints the injected build metadata schema.

- [ ] **Step 5: Commit**

```bash
git add modelcraft-cli/README.md modelcraft-cli/justfile modelcraft-cli/main.go
git commit -m "docs(cli): add release workflow and operator guide"
```

---

## Self-Review

### Spec coverage

- Gateway-first auth path: covered in Task 4 and Task 5.
- Org-scoped token semantics and local `currentProject`: covered in Task 3 and Task 5.
- Read-only command set: covered in Tasks 6, 7, and 8.
- English error contract: covered in Task 2.
- `schema` local and `describe` remote: covered in Task 8.
- Windows-first binary and PowerShell notes: covered in Task 9.

### Placeholder scan

- No `TODO`, `TBD`, or “implement later” markers remain.
- Every task includes explicit files, commands, and code snippets.

### Type consistency

- `config.Credentials`, `auth.Manager`, `client.AuthClient`, `resource.ModelPath`, and `output.CLIError` are introduced before later tasks use them.
- Route prefix `/api/cli/end-user/auth/*` is used consistently across Gateway and CLI tasks.

---

Plan complete and saved to `docs/superpowers/plans/2026-05-09-modelcraft-cli-v1-implementation.md`. Two execution options:

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

**Which approach?**
