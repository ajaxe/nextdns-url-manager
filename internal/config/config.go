package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

// Config represents the application configuration
type Config struct {
	Applications []Application `yaml:"applications"`
}

// Application represents an application group configuration
type Application struct {
	Name   string   `yaml:"name"`
	URLs   []string `yaml:"urls"`
	Enabled bool    `yaml:"enabled"`
	Timer  string   `yaml:"timer"`
}

// Load reads the configuration from a YAML file
func Load(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

// Save writes the configuration to a YAML file
func Save(config *Config, filePath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

// Merge updates the configuration with new application data and saves to the specified path
func Merge(existing *Config, newApplications []Application, filePath string) error {
	if existing == nil {
		existing = &Config{Applications: []Application{}}
	}

	for _, newApp := range newApplications {
		found := false
		for i := range existing.Applications {
			if existing.Applications[i].Name == newApp.Name {
				existing.Applications[i] = newApp
				found = true
				break
			}
		}
		if !found {
			existing.Applications = append(existing.Applications, newApp)
		}
	}
	return Save(existing, filePath)
}

// GetDefaultConfig returns a default configuration
func GetDefaultConfig() *Config {
	return &Config{
		Applications: []Application{
			{
				Name: "Entertainment",
				URLs: []string{"youtube.com", "netflix.com", "twitch.tv"},
				Enabled: false,
			},
			{
				Name: "Social Media",
				URLs: []string{"facebook.com", "twitter.com", "instagram.com", "tiktok.com"},
				Enabled: false,
			},
		},
	}
}
