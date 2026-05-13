package daemon

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"nextdns_client/internal/api"
	"nextdns_client/internal/config"
	"nextdns_client/internal/timer"

	"github.com/kardianos/service"
)

// Program implements service.Service and manages the daemon lifecycle.
type Program struct {
	apiKey     string
	profileID  string
	configPath string
	configFile string
	statePath  string
	cancel     context.CancelFunc
	done       chan struct{}
}

// Config returns the default service configuration.
func Config() *service.Config {
	return &service.Config{
		Name:        "nextdns-client",
		DisplayName: "NextDNS Client",
		Description: "Background service that periodically updates NextDNS deny/allow lists for enabled/disabled applications",
	}
}

func (p *Program) Start(s service.Service) error {
	p.done = make(chan struct{})

	// Initialize system logger
	logger, err := s.Logger(nil)
	if err == nil {
		InitServiceLogger(logger)
	}

	go func() {
		if err := p.run(); err != nil {
			slog.Error("daemon error", "error", err)
		}
	}()

	return nil
}

func (p *Program) run() error {
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	cfg, err := config.Load(p.configFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	apiClient := api.NewAPIClient(p.apiKey, p.profileID)

	// Auto-discover profile if not specified
	if p.profileID == "" {
		profiles, err := apiClient.ListProfiles()
		if err != nil {
			return fmt.Errorf("listing profiles: %w", err)
		}
		if len(profiles) == 0 {
			return fmt.Errorf("no profiles found")
		}
		p.profileID = profiles[0]["id"].(string)
		apiClient.SetProfileID(p.profileID)
	}

	api.SyncDisabledApps(apiClient, cfg)

	// Derive timer state path from config path
	stateDir := p.getConfigDir()
	timer.SetStateDir(stateDir)

	timer.StartDaemon(ctx, apiClient, cfg, p.configPath)

	cancel()
	close(p.done)
	return nil
}

func (p *Program) Stop(s service.Service) error {
	if p.cancel != nil {
		p.cancel()
	}

	select {
	case <-p.done:
		return nil
	case <-time.After(5 * time.Second):
		return nil
	}
}

// NewProgram creates a new Program with the given parameters.
// apiKey and profileID are sent as service arguments for lifecycle commands (install/start/stop).
// configPath is persisted to service.Config.Arguments for the daemon to read.
func NewProgram(apiKey, profileID, configPath string) (*Program, error) {
	program := &Program{
		configPath: configPath,
		configFile: configPath,
		done:       make(chan struct{}),
	}

	// Resolve to absolute path
	if abs, err := filepath.Abs(program.configPath); err == nil {
		program.configPath = abs
		program.configFile = abs
	}

	// For lifecycle commands invoked via service binary, read arguments from the service
	args := os.Args

	// Look for --config and --profile-id flags passed as service arguments
	for i := 0; i < len(args)-1; i++ {
		switch args[i] {
		case "--config", "-c":
			if i+1 < len(args) {
				program.configFile = args[i+1]
				program.configPath = args[i+1]
				i++
			}
		case "--profile-id", "-p":
			if i+1 < len(args) {
				program.profileID = args[i+1]
				i++
			}
		}
	}

	// For install/start/stop/status: allow setting credentials from constructor
	if apiKey != "" {
		program.apiKey = apiKey
	}
	if profileID != "" {
		program.profileID = profileID
	}

	if _, err := config.Load(program.configFile); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("loading config: %w", err)
		}
	}

	return program, nil
}

// SetServiceArgs stores arguments passed by the service command.
func (p *Program) SetServiceArgs(flags []string) {
	for i := 0; i < len(flags)-1; i++ {
		switch flags[i] {
		case "--config", "-c":
			if i+1 < len(flags) {
				p.configFile = flags[i+1]
				i++
			}
		case "--profile-id", "-p":
			if i+1 < len(flags) {
				p.profileID = flags[i+1]
				i++
			}
		}
	}
	p.configPath = p.configFile
}

// GetServiceConfigArgs extracts the service config Arguments for a Program.
// Returns flags for --api-key, --config, --profile-id.
func (p *Program) GetServiceConfigArgs() []string {
	var args []string
	if p.apiKey != "" {
		args = append(args, "--api-key", p.apiKey)
	}
	if p.configPath != "" {
		args = append(args, "--config", p.configPath)
	}
	if p.profileID != "" {
		args = append(args, "--profile-id", p.profileID)
	}
	return args
}

func (p *Program) getConfigDir() string {
	dir := filepath.Dir(p.configPath)
	if dir == "" {
		dir = "."
	}
	return dir
}

// Install creates a service and installs it.
func Install(svcConfig *service.Config) error {
	s, err := service.New(&Program{}, svcConfig)
	if err != nil {
		return err
	}
	return s.Install()
}

// StartService creates a service and starts it.
func StartService(svcConfig *service.Config) error {
	s, err := service.New(&Program{}, svcConfig)
	if err != nil {
		return err
	}
	return s.Start()
}

// StopService creates a service and stops it.
func StopService(svcConfig *service.Config) error {
	s, err := service.New(&Program{}, svcConfig)
	if err != nil {
		return err
	}
	return s.Stop()
}

// UninstallService creates a service and uninstalls it.
func UninstallService(svcConfig *service.Config) error {
	s, err := service.New(&Program{}, svcConfig)
	if err != nil {
		return err
	}
	return s.Uninstall()
}

// GetServiceStatus creates a service and returns its status.
func GetServiceStatus(svcConfig *service.Config) (service.Status, error) {
	s, err := service.New(&Program{}, svcConfig)
	if err != nil {
		return 0, err
	}
	return s.Status()
}

// --- Backward-compatible wrappers for daemon commands in root.go ---

// InstallProgram is a backward-compatible alias for Install.
func InstallProgram(program *Program) error {
	cfg := Config()
	return Install(cfg)
}

// StartProgram is a backward-compatible alias for StartService.
func StartProgram(program *Program) error {
	return StartService(Config())
}

// StopProgram is a backward-compatible alias for StopService.
func StopProgram(program *Program) error {
	return StopService(Config())
}

// UninstallProgram is a backward-compatible alias for UninstallService.
func UninstallProgram(program *Program) error {
	return UninstallService(Config())
}

// GetStatusProgram is a backward-compatible alias for GetServiceStatus.
func GetStatusProgram(program *Program) (service.Status, error) {
	return GetServiceStatus(Config())
}

// Wait waits for the program to complete.
func (p *Program) Wait() {
	<-p.done
}
