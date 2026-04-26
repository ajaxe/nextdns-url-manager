# Detailed Implementation Plan for NextDNS Client Application

## Overview
This document outlines a comprehensive implementation plan for a custom NextDNS client application that meets all requirements specified in the initial documentation. The application will be a cross-platform terminal UI application that interact with the NextDNS API, manage application groups with timer functionality, and persist configuration using YAML.

## Core Requirements Analysis

### 1. Cross-platform Terminal UI
- Requirements: Modern, intuitive layout with keyboard navigation (arrow keys)
- Framework: Bubble Tea TUI framework for robust cross-platform support
- Platform support: Windows and Linux compatibility

### 2. NextDNS API Integration
- Authentication: Client ID passed via CLI argument
- API Documentation: https://nextdns.github.io/api/
- Operations: Profile management and allowlist/denylist updates

### 3. Configuration Management
- Format: YAML configuration file
- Persistence: Local file system storage
- Merge capability: Sync changes made via UI with manual file updates
- Structure: Application groups with URLs, enabled status, and timer settings

### 4. Timer Functionality
- Time formats: `5s`, `70s`, `1m5s`, `1m 5s`, `1h5m`, `1h 5m`
- Background execution: Timer runs after UI exit
- Persistence: Store timer states for continuation

## Technical Implementation Plan

### Project Structure
```
nextdns_client/
├── main.go                 # Entry point
├── go.mod                  # Go module definition
├── go.sum                  # Go module checksums
├── cmd/                    # Command implementations
│   └── root.go             # Root command with CLI argument parsing  
├── internal/
│   ├── config/             # Configuration handling
│   │   ├── config.go       # Config structure and YAML loading/saving
│   │   └── config_test.go  # Config unit tests
│   ├── api/                # NextDNS API interactions
│   │   ├── api.go          # API client implementation
│   │   └── api_test.go     # API integration tests
│   ├── tui/                # Terminal UI components
│   │   ├── ui.go           # Main UI application and rendering
│   │   ├── app.go          # Application state management
│   │   └── components/     # UI component implementations
│   ├── timer/              # Timer functionality
│   │   ├── timer.go        # Timer parsing and execution logic
│   │   └── background.go   # Background timer executor with persistence
│   └── utils/              # Utility functions and helpers
├── docs/                   # Documentation including this plan
├── Makefile                # Build automation and deployment scripts
└── config.yaml             # Default configuration file template
```

### Phase 1: Project Setup and Core Framework
1. Initialize Go module with proper dependencies:
   - Cobra for CLI argument parsing
   - Bubble Tea for TUI framework
   - gopkg.in/yaml.v2 for YAML parsing
   - net/http for API interactions
2. Create basic project structure with directory organization
3. Implement CLI argument parsing for client_id

### Phase 2: Configuration System
1. Define YAML configuration structure for application groups:
   ```yaml
   applications:
     - name: "browser"
       urls:
         - "google.com"
         - "youtube.com"
       enabled: true
       timer: "30m"
     - name: "email"
       urls:
         - "gmail.com"
         - "outlook.com"
       enabled: false
       timer: null
   ```
2. Implement configuration loading, saving, and merging logic
3. Handle conflicts between manual file changes and application updates

### Phase 3: API Integration
1. Create NextDNS API client with authentication
2. Implement core API functionality:
   - `GET /profile`
   - `POST /allowlist` and `DELETE /allowlist`
   - `GET /status`

### Phase 4: Terminal UI Implementation
1. Design main UI layout with application groups
2. Implement keyboard navigation support
3. Create components for:
   - Application group listing
   - URL management
   - Enable/disable toggles
   - Timer configuration
4. Implement state management for UI interactions

### Phase 5: Timer Functionality
1. Timer format parsing system for:
   - `5s` through `70s` (seconds)
   - `1m5s` and `1m 5s` (minutes and seconds)
   - `1h5m` and `1h 5m` (hours and minutes)
2. Background timer execution with persistence
3. Graceful shutdown handling for timer continuation

### Phase 6: Testing and Validation
1. Unit tests for configuration and timer parsing
2. Integration tests for API interactions
3. UI testing for all interactive components
4. Cross-platform compatibility testing
5. End-to-end functional testing

## Technology Stack Details

### Go Packages and Libraries
1. **Cobra** - Command-line interface parsing
2. **Bubble Tea** - Terminal UI framework with modern look and feel
3. **gopkg.in/yaml.v2** - YAML parser for configuration files
4. **net/http or resty** - HTTP client for API interactions
5. **encoding/json** - For JSON API responses handling

### Implementation Components

#### CLI Interface (cmd/root.go)
- Parse client_id from command line argument
- Define application flags and help text
- Initialize main application components

#### Configuration Manager (internal/config/config.go)
- Load YAML configuration file
- Merge application changes with file contents
- Validate configuration structure
- Provide read/write access to application groups

#### API Client (internal/api/api.go)
- Authenticate with NextDNS using client_id
- Implement API endpoints:
  - `GET /profile`
  - `POST /allowlist` and `DELETE /allowlist`
  - `GET /status`

#### Timer System (internal/timer/timer.go)
- Parse timer syntax to Go duration
- Validate time formats
- Execute timer with background goroutines
- Handle timer events when application is not running

#### Terminal UI (internal/tui/ui.go)
- Main application state with Bubble Tea framework
- Keyboard navigation using arrow keys
- Interactive components for user management
- Real-time UI updates based on application state

## Cross-platform Compatibility
- All code developed for Go cross-platform compatibility
- Path handling using Go's `filepath` package
- Terminal UI behavior adapted for different platforms
- Build script for Windows (exe) and Linux (binary) outputs

## Security Considerations
- Client credentials passed only via CLI arguments
- No credential storage in configuration files
- Secure HTTP connections for API interactions
- Input validation for all user inputs

## Error Handling
- Comprehensive error logging with context
- Graceful degradation when API is unavailable
- User-friendly error messages in UI
- Resource cleanup on application exit

## Implementation Timeline
1. **Week 1**: Project setup, CLI, and basic configuration
2. **Week 2**: API integration and core functionality
3. **Week 3**: Terminal UI implementation with navigation
4. **Week 4**: Timer functionality and persistence
5. **Week 5**: Testing, validation, and documentation

This plan provides a comprehensive roadmap for implementing the NextDNS client application with all specified requirements, ensuring cross-platform compatibility, modern UI, and robust functionality for managing NextDNS application groups with timer features.