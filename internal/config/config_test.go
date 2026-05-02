package config

import (
	"os"
	"testing"
)

func TestLoadSave(t *testing.T) {
	tmpFile := "test_config.yaml"
	defer os.Remove(tmpFile)

	cfg := &Config{
		Applications: []Application{
			{
				Name:    "Test App",
				URLs:    []string{"example.com"},
				Enabled: true,
				Timer:   "1h",
			},
		},
	}

	err := Save(cfg, tmpFile)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	loaded, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(loaded.Applications) != 1 {
		t.Errorf("Expected 1 application, got %d", len(loaded.Applications))
	}

	if loaded.Applications[0].Name != "Test App" {
		t.Errorf("Expected 'Test App', got '%s'", loaded.Applications[0].Name)
	}
}

func TestMerge(t *testing.T) {
	tmpFile := "test_merge_config.yaml"
	defer os.Remove(tmpFile)

	existing := &Config{
		Applications: []Application{
			{
				Name: "App1",
				URLs: []string{"url1"},
			},
		},
	}

	newApps := []Application{
		{
			Name: "App1",
			URLs: []string{"url1_updated"},
		},
		{
			Name: "App2",
			URLs: []string{"url2"},
		},
	}

	err := Merge(existing, newApps, tmpFile)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	loaded, _ := Load(tmpFile)
	if len(loaded.Applications) != 2 {
		t.Errorf("Expected 2 applications after merge, got %d", len(loaded.Applications))
	}

	for _, app := range loaded.Applications {
		if app.Name == "App1" && app.URLs[0] != "url1_updated" {
			t.Errorf("App1 was not updated correctly")
		}
	}
}
