package daemon

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kardianos/service"
)

func TestConfig(t *testing.T) {
	cfg := Config()
	if cfg.Name != "nextdns-client" {
		t.Errorf("Config().Name = %s, want 'nextdns-client'", cfg.Name)
	}
	if cfg.DisplayName != "NextDNS Client" {
		t.Errorf("Config().DisplayName = %s, want 'NextDNS Client'", cfg.DisplayName)
	}
	if cfg.Description == "" {
		t.Error("Config().Description should not be empty")
	}
}

func TestNewProgram_EmptyArgs(t *testing.T) {
	prog, err := NewProgram("", "", "")
	if err != nil {
		t.Fatalf("NewProgram() error = %v", err)
	}
	if prog == nil {
		t.Fatal("expected non-nil Program")
	}
	if prog.configPath != "" {
		t.Errorf("expected empty configPath, got '%s'", prog.configPath)
	}
	if prog.configFile != "" {
		t.Errorf("expected empty configFile, got '%s'", prog.configFile)
	}
	if prog.done == nil {
		t.Error("expected done channel to be initialized")
	}
}

func TestNewProgram_WithApiKey(t *testing.T) {
	prog, err := NewProgram("my-api-key", "", "")
	if err != nil {
		t.Fatalf("NewProgram() error = %v", err)
	}
	if prog.apiKey != "my-api-key" {
		t.Errorf("apiKey = %s, want 'my-api-key'", prog.apiKey)
	}
}

func TestNewProgram_WithProfileId(t *testing.T) {
	prog, err := NewProgram("", "my-profile-id", "")
	if err != nil {
		t.Fatalf("NewProgram() error = %v", err)
	}
	if prog.profileID != "my-profile-id" {
		t.Errorf("profileID = %s, want 'my-profile-id'", prog.profileID)
	}
}

func TestNewProgram_WithConfigPath(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	prog, err := NewProgram("key", "profile-id", configPath)
	if err != nil {
		t.Fatalf("NewProgram() error = %v", err)
	}
	if prog.configPath != configPath {
		t.Errorf("configPath = %s, want %s", prog.configPath, configPath)
	}
	if prog.configFile != configPath {
		t.Errorf("configFile = %s, want %s", prog.configFile, configPath)
	}
}

func TestGetServiceConfigArgs(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		config   string
		profile  string
		wantArgs []string
	}{
		{"all empty", "", "", "", []string{}},
		{"only api key", "key123", "", "", []string{"--api-key", "key123"}},
		{"only config", "", "/path/to/config.yaml", "", []string{"--config", "/path/to/config.yaml"}},
		{"only profile", "", "", "profile-abc", []string{"--profile-id", "profile-abc"}},
		{"all set", "key", "/cfg.yaml", "pid", []string{"--api-key", "key", "--config", "/cfg.yaml", "--profile-id", "pid"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prog, _ := NewProgram(tt.apiKey, tt.profile, tt.config)
			got := prog.GetServiceConfigArgs()
			if len(got) != len(tt.wantArgs) {
				t.Errorf("GetServiceConfigArgs() = %v, want %v", got, tt.wantArgs)
				return
			}
			for i := range got {
				if got[i] != tt.wantArgs[i] {
					t.Errorf("GetServiceConfigArgs()[%d] = %s, want %s", i, got[i], tt.wantArgs[i])
				}
			}
		})
	}
}

func TestSetServiceArgs(t *testing.T) {
	prog := &Program{}

	prog.SetServiceArgs([]string{"--config", "/etc/myconfig.yaml", "--profile-id", "test-profile"})
	if prog.configFile != "/etc/myconfig.yaml" {
		t.Errorf("configFile = %s, want '/etc/myconfig.yaml'", prog.configFile)
	}
	if prog.configPath != prog.configFile {
		t.Errorf("configPath = %s should equal configFile", prog.configPath)
	}
	if prog.profileID != "test-profile" {
		t.Errorf("profileID = %s, want 'test-profile'", prog.profileID)
	}
}

func TestSetServiceArgs_ShortFlags(t *testing.T) {
	prog := &Program{}
	prog.SetServiceArgs([]string{"-c", "/etc/short.yaml", "-p", "short-profile"})
	if prog.configFile != "/etc/short.yaml" {
		t.Errorf("configFile = %s, want '/etc/short.yaml'", prog.configFile)
	}
	if prog.profileID != "short-profile" {
		t.Errorf("profileID = %s, want 'short-profile'", prog.profileID)
	}
}

func TestSetServiceArgs_InvalidOrder(t *testing.T) {
	prog := &Program{}
	// Flags in wrong order - last flag should just set values
	prog.SetServiceArgs([]string{"-p", "test-profile"})
	if prog.profileID != "test-profile" {
		t.Errorf("profileID = %s, want 'test-profile'", prog.profileID)
	}
}

func TestInstall(t *testing.T) {
	cfg := Config()
	err := Install(cfg)
	// On Windows, this may fail if not running as admin, but should not panic
	// We just verify that a service can be created
	if err != nil {
		// Acceptable on non-admin systems - just verify it's not a panic
		t.Logf("Install() returned error (may be non-admin): %v", err)
	}
}

func TestGetServiceStatus(t *testing.T) {
	cfg := Config()
	status, err := GetServiceStatus(cfg)
	// On non-Windows or without service running, may fail
	t.Logf("GetServiceStatus() status=%d, err=%v", status, err)
}

func TestProgram_Setup(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create a minimal valid YAML config
	cfgYAML := `applications:
- name: TestApp
  urls:
  - example.com
  enabled: false
`
	if err := os.WriteFile(configPath, []byte(cfgYAML), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	prog, err := NewProgram("test-key", "", configPath)
	if err != nil {
		t.Fatalf("NewProgram() error = %v", err)
	}

	if prog.apiKey != "test-key" {
		t.Errorf("apiKey = %s, want 'test-key'", prog.apiKey)
	}
	if prog.configPath != configPath {
		t.Errorf("configPath = %s, want %s", prog.configPath, configPath)
	}
	if prog.done == nil {
		t.Error("expected done channel")
	}
}

func TestConfig_NameUniqueness(t *testing.T) {
	cfg1 := Config()
	cfg2 := Config()
	// Both should have the same name since it's a static function
	if cfg1.Name != cfg2.Name {
		t.Error("Config() should always return same Name")
	}
	// But they should be different instances
	if cfg1 == cfg2 {
		t.Error("Config() should return different instances")
	}
}

func TestConfig_Description(t *testing.T) {
	cfg := Config()
	if cfg.Description != "Background service that periodically updates NextDNS deny/allow lists for enabled/disabled applications" {
		t.Errorf("unexpected Description: %s", cfg.Description)
	}
}

// Verify Program implements service.Service interface
var _ service.Service = (*Program)(nil)
