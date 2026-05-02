package api

import (
	"sync"

	"nextdns_client/internal/config"
)

// SyncDisabledApps synchronizes applications that are disabled in the config with NextDNS denylist
func SyncDisabledApps(client *APIClient, cfg *config.Config) {
	var wg sync.WaitGroup

	for _, app := range cfg.Applications {
		if !app.Enabled {
			for _, url := range app.URLs {
				wg.Add(1)
				go func(u string) {
					defer wg.Done()
					// Optimistic addition: ignore errors (e.g. if already in denylist)
					_ = client.AddToDenylist(u)
				}(url)
			}
		}
	}

	wg.Wait()
}
