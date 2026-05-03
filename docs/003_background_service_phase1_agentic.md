# Phase 1: Background Service Wrapper (`internal/daemon`)

## Overview
This phase focuses on wrapping the existing `timer.StartDaemon` logic into a structure compatible with `github.com/kardianos/service`. This allows the application to behave as a native service on Windows and Linux while maintaining its core logic.

## Implementation Details

### 1. New Package: `internal/daemon`
Create a new file `internal/daemon/service.go`. This package will handle the service lifecycle and abstraction.

### 2. The `program` Struct
The `program` struct will implement `service.Interface`. It acts as the bridge between the OS service manager and our application logic.

```go
type program struct {
    apiKey     string
    profileID  string
    configPath string
    exit       chan struct{}
    cancel     context.CancelFunc
}
```

### 3. Lifecycle Methods

#### `Start(s service.Service) error`
- **Goal**: Non-blocking initialization.
- **Actions**:
    - Initialize `exit` channel.
    - Start the `run()` method in a background goroutine.
    - Return `nil` immediately.

#### `run()` (Internal)
- **Goal**: The actual execution loop.
- **Actions**:
    - Create a cancellable context (`context.WithCancel`). Store `cancel` in the `program` struct.
    - Load configuration via `config.Load(p.configPath)`.
    - Initialize `api.NewAPIClient(p.apiKey, p.profileID)`.
    - Handle profile discovery if `profileID` is empty.
    - Execute `api.SyncDisabledApps`.
    - Call `timer.StartDaemon(ctx, apiClient, cfg, p.configPath)`.
    - Once `StartDaemon` returns (due to context cancellation), close the `exit` channel or signal completion.

#### `Stop(s service.Service) error`
- **Goal**: Graceful shutdown.
- **Actions**:
    - Call the stored `cancel()` function.
    - (Optional) Wait for a short duration or until `exit` is signaled to ensure `StartDaemon` has cleaned up.
    - Return `nil`.

## Task List

- [ ] **Dependency**: Add `github.com/kardianos/service` to `go.mod`.
- [ ] **Infrastructure**: Create `internal/daemon` directory.
- [ ] **Code**: Implement `internal/daemon/service.go`.
    - [ ] Define `program` struct.
    - [ ] Implement `Start` (async execution).
    - [ ] Implement `Stop` (graceful cancellation).
    - [ ] Implement `GetConfig` helper to create `service.Config`.
- [ ] **Integration**: Prepare a temporary test in `cmd/root.go` or a separate test file to verify the `program` struct can be initialized.

## Verification Plan
1. **Unit Test**: Mock `timer.StartDaemon` (if possible) or use a short-lived context to verify `Start` and `Stop` trigger the correct internal state changes.
2. **Manual Integration**: Verify that calling `Start` correctly initializes the API client and begins the periodic check loop.
