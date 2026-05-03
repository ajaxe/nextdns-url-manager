# Phase 2: CLI Refactoring (`cmd/service.go`)

## Overview
This phase introduces the `service` command to the CLI, providing subcommands to manage the application as an OS-native service. It leverages the `internal/daemon` package (implemented in Phase 1) and the `github.com/kardianos/service` library for cross-platform service management.

## Implementation Details

### 1. New File: `cmd/service.go`
Create a new file `cmd/service.go` to define the `service` command and its subcommands.

### 2. Command Structure

```bash
nextdns-client service
├── install    # Registers the service with the OS
├── uninstall  # Removes the service registration
├── start      # Starts the registered service
├── stop       # Stops the registered service
└── run        # Entry point for the OS service manager (internal)
```

### 3. Subcommand Logic

#### `service install`
- **Goal**: Persist configuration and register the binary.
- **Flags**: `--api-key`, `--profile-id`, `--config`.
- **Logic**:
    - Validate required flags (especially `--api-key`).
    - Resolve the absolute path of the configuration file if a relative path is provided.
    - Initialize `service.Config`.
    - Set `Arguments` to: `["service", "run", "--api-key", "...", "--profile-id", "...", "--config", "..."]`.
    - Call `s.Install()`.
    - Output success or error message.

#### `service run`
- **Goal**: Execute the daemon logic within the service context.
- **Flags**: Inherits or mirrors flags from `install`.
- **Logic**:
    - This command is intended to be called by the OS Service Control Manager.
    - Initialize the `program` struct (from `internal/daemon`) using the provided flags.
    - Create a `service.Service` object.
    - Call `s.Run()`. This is a blocking call that enters the service loop.

#### `service start / stop / uninstall`
- **Goal**: Control the existing service.
- **Logic**:
    - Initialize a minimal `service.Config` (only needs the service name).
    - Call `service.Control(s, action)`.

### 4. Integration with `cmd/root.go`
- Register `serviceCmd` to `rootCmd` in `init()`.
- Ensure flag consistency between `rootCmd`, `daemonCmd`, and `serviceCmd`.

## Task List

- [ ] **Infrastructure**: Create `cmd/service.go`. (priority: high)
- [ ] **Code**: Implement `serviceCmd` and subcommands. (priority: high)
    - [ ] Implement `install` with flag persistence in `Arguments`. (priority: high)
    - [ ] Implement `run` calling `s.Run()`. (priority: high)
    - [ ] Implement `start`, `stop`, `uninstall` using `service.Control`. (priority: high)
- [ ] **Refinement**: (priority: high)
    - [ ] Ensure absolute path resolution for `--config`. (priority: high)
    - [ ] Add descriptive help text for each subcommand. (priority: high)
- [ ] **Cleanup**: (Optional) Decide whether to deprecate `daemonCmd` or keep it for foreground debugging. (priority: high)

## Verification Plan

1. **CLI Help**: Run `nextdns-client service --help` to verify the command structure.
2. **Install Simulation**: Run `nextdns-client service install --api-key TEST --profile-id PROFIL --config config.yaml` and verify it attempts to register the service with the correct arguments (using a dry-run or checking logs).
3. **Execution**: Verify that `nextdns-client service run` (when called manually) correctly starts the daemon logic (it should stay running and perform periodic checks).
4. **Platform Testing**:
    - **Windows**: Run `install` (as Admin) and verify service appears in `services.msc`. Run `start` and verify it moves to "Running" state.
    - **Linux**: Run `install` (as root/sudo) and verify `/etc/systemd/system/nextdns-client.service` is created.
