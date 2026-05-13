# NextDNS URL Manager

A cross-platform terminal UI application for managing NextDNS application groups with timer functionality and background service support.

## Features

- **Modern Terminal UI** — Built with Bubble Tea and Lip Gloss; keyboard navigation with arrow keys and shortcuts
- **NextDNS API Integration** — Connect via `--api-key` CLI flag; manage profiles and allowlists
- **Application Groups** — Create and manage URL groups with enable/disable toggles
- **Timer Functionality** — Schedule periods (`5s`, `1m5s`, `1h5m`, etc.) to temporarily disable a group; timers persist across restarts
- **Hot-Reload** — Background daemon detects config file changes in real time without restart
- **Cross-Platform Service** — Install as a native Windows Service or systemd service; starts on boot
- **YAML Configuration** — Persistent config at `config.yaml` with merge support for manual edits

## Project Structure

```
.
├── cmd/
│   ├── main.go            # Entry point
│   ├── root.go            # Root command and CLI argument parsing
│   └── service.go         # Service install/start/stop/uninstall subcommands
├── internal/
│   ├── api/               # NextDNS API client
│   ├── config/            # YAML config loading, saving, and merging
│   ├── daemon/            # Cross-platform service wrapper (kardianos/service)
│   │   ├── service.go     # Service lifecycle (Start, Stop)
│   │   └── logging.go     # slog → Windows Event Log / Linux journald bridge
│   ├── timer/             # Timer parsing and background daemon
│   │   ├── background.go  # Background check loop with hot-reload
│   │   └── timer.go       # Duration parsing and timer state persistence
│   └── tui/               # Bubble Tea terminal UI
├── docs/                  # Architecture Decision Records & implementation plans
│   ├── 001_initial_requirements_agentic.md
│   ├── 002_add_app_group_with_hotreload_agentic.md
│   ├── 003_background_service_agentic.md
│   ├── 003_background_service_phase1_agentic.md
│   ├── 003_background_service_phase2_agentic.md
│   └── 003_background_service_phase3_agentic.md
├── config.yaml            # Default configuration template
├── Makefile
└── go.mod
```

## Commands

```bash
# Show help
nextdns-client help

# Start TUI (requires --api-key; --profile-id is auto-discovered if omitted)
nextdns-client --api-key YOUR_API_KEY

# Service management
nextdns-client service install --api-key YOUR_API_KEY --config ./config.yaml [--display-name "My Service"]
nextdns-client service start
nextdns-client service stop
nextdns-client service status
nextdns-client service uninstall

# Legacy daemon subcommands
nextdns-client daemon run   --api-key KEY --config ./config.yaml
nextdns-client daemon install --api-key KEY --config ./config.yaml
nextdns-client daemon start
nextdns-client daemon stop
nextdns-client daemon status
nextdns-client daemon uninstall
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--api-key` | `-k` | *(required)* | API key for NextDNS authentication |
| `--profile-id` | `-p` | auto-discovered | NextDNS profile ID |
| `--config` | `-c` | `config.yaml` | Path to configuration file |
| `--debug` | `-d` | `false` | Enable debug mode (shows API logs in TUI) |
| `--display-name` | — | service name | Display name for the installed service |

## Configuration

The application reads `config.yaml` by default:

```yaml
applications:
  - name: Entertainment
    urls:
      - twitch.tv
    enabled: false
    timer: ""
  - name: Social Media
    urls:
      - facebook.com
      - twitter.com
    enabled: false
    timer: ""
```

- **name** — Display name for the application group
- **urls** — List of URLs/domains in this group
- **enabled** — Whether the group is currently active
- **timer** — Duration string (e.g., `1m`, `30m`, `1h5m`) for timed enablement

### Timer Formats

| Format | Meaning |
|--------|---------|
| `5s`   | 5 seconds |
| `70s`  | 70 seconds |
| `1m5s` | 1 minute 5 seconds |
| `1m 5s`| 1 minute 5 seconds |
| `1h5m` | 1 hour 5 minutes |
| `1h 5m`| 1 hour 5 minutes |

## Build

### Prerequisites

- Go 1.26+
- PowerShell (for `make build`)

### Quick Start

```bash
# Build
make build

# Run the TUI
make run

# Run tests
make test

# Lint
make lint
```

### Cross-Platform Builds

```bash
make dist                # Build all platforms
make win64               # Windows x64
make linux64             # Linux x64
make macarm64            # macOS ARM64
```

## Architecture Notes

Key design decisions are captured as Architecture Decision Records (ADRs) in `docs/`:

- **ADR-001**: [Initial Requirements & Implementation Plan](docs/001_initial_requirements_agentic.md) — Technology choices (Bubble Tea, Cobra, kardianos/service), phased implementation plan, tech stack
- **ADR-002**: [Application Groups with Hot-Reload](docs/002_add_app_group_with_hotreload_agentic.md) — TUI state machine for group creation, daemon hot-reload via config polling
- **ADR-003**: [Cross-Platform Background Service](docs/003_background_service_agentic.md) — kardianos/service library, `service` CLI subcommand, service lifecycle
- **ADR-003 Phase 1**: [Service Wrapper](docs/003_background_service_phase1_agentic.md) — `internal/daemon` package, `program` struct with Start/Stop, goroutine-based daemon
- **ADR-003 Phase 2**: [CLI Refactoring](docs/003_background_service_phase2_agentic.md) — `nextdns-client service` subcommand with install/start/stop/uninstall/run
- **ADR-003 Phase 3**: [Logging & Persistence](docs/003_background_service_phase3_agentic.md) — slog → Event Log/journald bridge, absolute path resolution, systemd service config

## Security

- `--api-key` is passed via CLI arguments only — never hardcoded
- Service runs as `SYSTEM` (Windows) or `root` (Linux) for file access permissions
- Configuration paths are resolved to absolute paths before service installation

## Troubleshooting

### Service not starting on Windows
- Run with elevated (Admin) privileges: `nextdns-client service install`
- Check Event Viewer → Windows Logs → Application for entries from "NextDNS Client"

### Service not starting on Linux
- Ensure `sudo` is used for install/start
- Check logs: `journalctl -u nextdns-client -f`

### Config file not found
- Service runs from `C:\Windows\System32` (Windows) or `/` (Linux). Always use absolute paths:
   ```bash
   nextdns-client service install --api-key KEY --config "C:\full\path\to\config.yaml"
   ```
