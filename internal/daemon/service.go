package daemon

import (
	"context"
	"fmt"
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
	apiKey      string
	profileID   string
	configPath  string
	cancel      context.CancelFunc
	done        chan struct{}
	configFile  string
	configStore map[string]string
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

	go func() {
		if err := p.run(); err != nil {
			fmt.Fprintf(os.Stderr, "daemon error: %v\n", err)
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
		fmt.Printf("Using profile: %s (%s)\n", profiles[0]["name"], p.profileID)
	}

	api.SyncDisabledApps(apiClient, cfg)

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
		configPath:  configPath,
		configFile:  configPath,
		configStore: make(map[string]string),
		done:        make(chan struct{}),
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
				p.configPath = flags[i+1]
				i++
			}
		case "--profile-id", "-p":
			if i+1 < len(flags) {
				p.profileID = flags[i+1]
				i++
			}
		}
	}
}

// GetServiceConfigArgs extracts the service config Arguments for a Program.
// Returns flags for --config and --profile-id (non-sensitive path/ID only).
func (p *Program) GetServiceConfigArgs() []string {
	var args []string
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
func Install(program *Program) error {
	cfg := Config()

	var args []string
	if len(os.Args) > 2 {
		args = []string{filepath.Base(os.Args[0]), "--config", program.configPath}
		cfg.DisplayName = os.Args[2]
	}

	cfg.Arguments = args

	s, err := service.New(program, cfg)
	if err != nil {
		return err
	}
	return s.Install()
}

// Start creates a service and starts it.
func Start(program *Program) error {
	s, err := service.New(program, Config())
	if err != nil {
		return err
	}
	return s.Start()
}

// Stop creates a service and stops it.
func Stop(program *Program) error {
	s, err := service.New(program, Config())
	if err != nil {
		return err
	}
	return s.Stop()
}

// Uninstall creates a service and uninstalls it.
func Uninstall(program *Program) error {
	s, err := service.New(program, Config())
	if err != nil {
		return err
	}
	return s.Uninstall()
}

// GetStatus creates a service and returns its status.
func GetStatus(program *Program) (service.Status, error) {
	s, err := service.New(program, Config())
	if err != nil {
		return 0, err
	}
	return s.Status()
}

// Wait waits for the program to complete.
func (p *Program) Wait() {
	<-p.done
}
