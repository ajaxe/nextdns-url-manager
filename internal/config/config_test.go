package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_Save_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "config.yaml")

	original := &Config{
		Applications: []Application{
			{Name: "Test App", URLs: []string{"example.com"}, Enabled: true, Timer: "1h"},
			{Name: "Another", URLs: []string{"a.com", "b.com"}, Enabled: false, Timer: "30m"},
		},
	}

	if err := Save(original, filePath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := Load(filePath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loaded.Applications) != 2 {
		t.Fatalf("expected 2 applications, got %d", len(loaded.Applications))
	}
	if loaded.Applications[0].Name != "Test App" {
		t.Errorf("first app name = %s, want 'Test App'", loaded.Applications[0].Name)
	}
	if loaded.Applications[0].Enabled != true {
		t.Errorf("first app enabled = %v, want true", loaded.Applications[0].Enabled)
	}
	if loaded.Applications[1].Timer != "30m" {
		t.Errorf("second app timer = %s, want '30m'", loaded.Applications[1].Timer)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "nonexistent.yaml"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "empty.yaml")
	if err := os.WriteFile(filePath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	cfg, err := Load(filePath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if len(cfg.Applications) != 0 {
		t.Errorf("expected empty applications, got %d", len(cfg.Applications))
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "invalid.yaml")
	if err := os.WriteFile(filePath, []byte("{invalid yaml [[[["), 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	_, err := Load(filePath)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestSave_EmptyConfig(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "empty_config.yaml")

	empty := &Config{Applications: []Application{}}
	if err := Save(empty, filePath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty YAML output")
	}
}

func TestLoad_MultipleApps(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "multi.yaml")

	yamlContent := `applications:
- name: App1
  urls:
  - url1.com
  enabled: true
- name: App2
  urls:
  - url2.com
  - url3.com
  enabled: false
- name: App3
  urls:
  - url4.com
  timer: 2h
`
	if err := os.WriteFile(filePath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	cfg, err := Load(filePath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.Applications) != 3 {
		t.Fatalf("expected 3 applications, got %d", len(cfg.Applications))
	}
	if cfg.Applications[1].Name != "App2" {
		t.Errorf("second app name = %s, want 'App2'", cfg.Applications[1].Name)
	}
	if len(cfg.Applications[1].URLs) != 2 {
		t.Errorf("second app URLs count = %d, want 2", len(cfg.Applications[1].URLs))
	}
}

func TestMerge_AppendNew(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "merge_append.yaml")

	existing := &Config{
		Applications: []Application{
			{Name: "App1", URLs: []string{"url1"}, Enabled: true},
		},
	}

	newApps := []Application{
		{Name: "App2", URLs: []string{"url2"}, Enabled: false},
	}

	if err := Merge(existing, newApps, filePath); err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	loaded, err := Load(filePath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loaded.Applications) != 2 {
		t.Fatalf("expected 2 apps after merge, got %d", len(loaded.Applications))
	}
}

func TestMerge_UpdateExisting(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "merge_update.yaml")

	existing := &Config{
		Applications: []Application{
			{Name: "App1", URLs: []string{"url1"}, Enabled: true},
		},
	}

	newApps := []Application{
		{Name: "App1", URLs: []string{"url1_updated"}, Enabled: false},
	}

	if err := Merge(existing, newApps, filePath); err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	loaded, err := Load(filePath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loaded.Applications) != 1 {
		t.Fatalf("expected 1 app (updated), got %d", len(loaded.Applications))
	}
	if loaded.Applications[0].Enabled != false {
		t.Errorf("App1 should be updated to disabled")
	}
	if loaded.Applications[0].URLs[0] != "url1_updated" {
		t.Errorf("App1 URLs should be updated, got %v", loaded.Applications[0].URLs)
	}
}

func TestMerge_NilConfig(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "merge_nil.yaml")

	var existing *Config
	newApps := []Application{
		{Name: "App1", URLs: []string{"url1"}, Enabled: true},
	}

	if err := Merge(existing, newApps, filePath); err != nil {
		t.Fatalf("Merge(nil, ...) error = %v", err)
	}

	loaded, err := Load(filePath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loaded.Applications) != 1 {
		t.Fatalf("expected 1 app from nil merge, got %d", len(loaded.Applications))
	}
}

func TestMerge_MultipleUpdateAppend(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "merge_mixed.yaml")

	existing := &Config{
		Applications: []Application{
			{Name: "Existing", URLs: []string{"url1"}, Enabled: true},
		},
	}

	newApps := []Application{
		{Name: "Existing", URLs: []string{"url1_updated"}, Enabled: false},  // update
		{Name: "New", URLs: []string{"url2"}, Enabled: true},                // append
		{Name: "AnotherNew", URLs: []string{"url3"}, Enabled: true},          // append
	}

	if err := Merge(existing, newApps, filePath); err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	loaded, err := Load(filePath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loaded.Applications) != 3 {
		t.Fatalf("expected 3 apps, got %d", len(loaded.Applications))
	}
}

func TestGetDefaultConfig(t *testing.T) {
	cfg := GetDefaultConfig()
	if cfg == nil {
		t.Fatal("expected non-nil default config")
	}
	if len(cfg.Applications) != 2 {
		t.Fatalf("expected 2 default apps, got %d", len(cfg.Applications))
	}

	// Check Entertainment
	ent := cfg.Applications[0]
	if ent.Name != "Entertainment" {
		t.Errorf("first app name = %s, want 'Entertainment'", ent.Name)
	}
	if ent.Enabled != false {
		t.Errorf("Entertainment enabled = %v, want false", ent.Enabled)
	}
	if len(ent.URLs) != 3 {
		t.Errorf("Entertainment URLs = %d, want 3", len(ent.URLs))
	}

	// Check Social Media
	social := cfg.Applications[1]
	if social.Name != "Social Media" {
		t.Errorf("second app name = %s, want 'Social Media'", social.Name)
	}
	if social.Enabled != false {
		t.Errorf("Social Media enabled = %v, want false", social.Enabled)
	}
	if len(social.URLs) != 4 {
		t.Errorf("Social Media URLs = %d, want 4", len(social.URLs))
	}
}

func TestSave_Load_SpecialChars(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "specialchars.yaml")

	cfg := &Config{
		Applications: []Application{
			{Name: "App with spaces & special!@#", URLs: []string{"test.com/path?query=value"}, Enabled: true},
		},
	}

	if err := Save(cfg, filePath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := Load(filePath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.Applications[0].Name != "App with spaces & special!@#" {
		t.Errorf("name = %s, want 'App with spaces & special!@#'", loaded.Applications[0].Name)
	}
}

func TestSave_Load_PreservesOrder(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "order.yaml")

	cfg := &Config{
		Applications: []Application{
			{Name: "First", URLs: []string{"url1"}, Enabled: true},
			{Name: "Second", URLs: []string{"url2"}, Enabled: false},
			{Name: "Third", URLs: []string{"url3"}, Enabled: true},
		},
	}

	if err := Save(cfg, filePath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := Load(filePath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loaded.Applications) != 3 {
		t.Fatalf("expected 3 apps, got %d", len(loaded.Applications))
	}
	if loaded.Applications[0].Name != "First" || loaded.Applications[1].Name != "Second" || loaded.Applications[2].Name != "Third" {
		t.Error("application order should be preserved")
	}
}

func TestLoad_YAMLWithExtraFields(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "extra.yaml")

	yamlContent := `applications:
- name: App1
  urls:
  - url1.com
  enabled: true
  extra_field: should_be_ignored
other_section:
  nested: data
`
	if err := os.WriteFile(filePath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	cfg, err := Load(filePath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.Applications) != 1 {
		t.Fatalf("expected 1 app, got %d", len(cfg.Applications))
	}
}
