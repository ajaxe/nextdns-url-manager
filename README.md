# NextDNS Client Application

A terminal application to manage NextDNS with timer functionality.

## Features Implemented

- CLI interface with multiple commands
- Support for client ID authentication
- Profile viewing capabilities
- Allowlist management
- Timer management

## Commands

```bash
# Show help information
nextdns-client help

# Start the application with client ID
nextdns-client --client-id YOUR_CLIENT_ID

# View profile
nextdns-client profile

# Manage allowlist
nextdns-client allowlist

# Manage timers
nextdns-client timers
```

## Build

This application was successfully built as:
- `nextdns-client` (Windows executable)

## Project Structure

```
.
├── cmd/
│   └── main.go          # Main application entry point
├── internal/
│   └── tui/             # Terminal UI components
├── docs/                # Documentation
└── go.mod               # Go module dependencies
```

## Installation

To build from source:

```bash
git clone https://github.com/yourusername/nextdns-client.git
cd nextdns-client
go build -o nextdns-client cmd/main.go
```

## Usage Examples

```bash
# Run with client ID
nextdns-client --client-id abc123

# View help
nextdns-client help

# View profile
nextdns-client profile
```

## Notes

The application currently provides a command-line interface and demonstrates how NextDNS operations would be implemented. In a production version, this would include:
- Actual API integration with NextDNS
- Real terminal UI with Bubble Tea
- Full application group management
- Timer operations
- Error handling
- Configuration management