package timer

import (
	"context"
	"fmt"
	"time"

	"nextdns_client/internal/api"
	"nextdns_client/internal/config"
)

// RunBackgroundCheck checks all active timers and updates NextDNS if they expired
func RunBackgroundCheck(apiClient *api.APIClient, cfg *config.Config, configPath string) error {
	state, err := LoadState()
	if err != nil {
		return fmt.Errorf("failed to load timer state: %w", err)
	}

	newTimers := []Timer{}
	changed := false

	for _, t := range state.Timers {
		if !t.Running {
			continue
		}

		if t.GetRemainingTime() <= 0 {
			// Timer expired!
			fmt.Printf("[%s] Timer expired for %s. Blocking application group.\n", time.Now().Format(time.Kitchen), t.Name)
			
			// Find the app in config
			for i := range cfg.Applications {
				if cfg.Applications[i].Name == t.TargetApp {
					cfg.Applications[i].Enabled = false
					
					// Update NextDNS: Block the URLs now that the "allow" period is over
					for _, url := range cfg.Applications[i].URLs {
						err := apiClient.AddToDenylist(url)
						if err != nil {
							fmt.Printf("Error updating NextDNS for %s: %v\n", url, err)
						}
					}
					changed = true
					break
				}
			}
		} else {
			newTimers = append(newTimers, t)
		}
	}

	if changed {
		state.Timers = newTimers
		SaveState(state)
		config.Save(cfg, configPath)
	}

	return nil
}

// StartDaemon runs the background check periodically until the context is cancelled
func StartDaemon(ctx context.Context, apiClient *api.APIClient, cfg *config.Config, configPath string) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Initial check on start
	RunBackgroundCheck(apiClient, cfg, configPath)

	for {
		select {
		case <-ticker.C:
			if err := RunBackgroundCheck(apiClient, cfg, configPath); err != nil {
				fmt.Printf("Background check error: %v\n", err)
			}
		case <-ctx.Done():
			fmt.Println("Daemon stopping...")
			return
		}
	}
}
