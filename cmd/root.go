package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"nextdns_client/internal/api"
	"nextdns_client/internal/config"
	"nextdns_client/internal/timer"
	"nextdns_client/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nextdns-client",
	Short: "NextDNS Client Application",
	Long:  "A powerful terminal application to manage NextDNS application groups with timer functionality",

	Run: func(cmd *cobra.Command, args []string) {
		apiKey, _ := cmd.Flags().GetString("api-key")
		profileID, _ := cmd.Flags().GetString("profile-id")
		configPath, _ := cmd.Flags().GetString("config")
		debug, _ := cmd.Flags().GetBool("debug")

		if apiKey == "" {
			fmt.Println("Error: --api-key is required")
			os.Exit(1)
		}

		// Load or create config
		cfg, err := config.Load(configPath)
		if err != nil {
			if os.IsNotExist(err) {
				cfg = config.GetDefaultConfig()
				config.Save(cfg, configPath)
			} else {
				fmt.Printf("Error loading config: %v\n", err)
				os.Exit(1)
			}
		}

		// Initialize API client
		apiClient := api.NewAPIClient(apiKey, profileID)

		// If profileID is empty, try to fetch the first one
		if profileID == "" {
			profiles, err := apiClient.ListProfiles()
			if err != nil || len(profiles) == 0 {
				fmt.Println("Error: --profile-id is required or no profiles found")
				if err != nil {
					fmt.Printf("API Error: %v\n", err)
				}
				os.Exit(1)
			}
			profileID = profiles[0]["id"].(string)
			apiClient.SetProfileID(profileID)
			fmt.Printf("Using profile: %s (%s)\n", profiles[0]["name"], profileID)
		}

		// Sync disabled apps to denylist on startup
		api.SyncDisabledApps(apiClient, cfg)

		// One-time background check on startup
		timer.RunBackgroundCheck(apiClient, cfg, configPath)

		p := tea.NewProgram(tui.NewModel(apiKey, profileID, cfg, apiClient, configPath, debug))
		if _, err := p.Run(); err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
			os.Exit(1)
		}
	},
}

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Run the background timer daemon",
	Run: func(cmd *cobra.Command, args []string) {
		apiKey, _ := cmd.Flags().GetString("api-key")
		profileID, _ := cmd.Flags().GetString("profile-id")
		configPath, _ := cmd.Flags().GetString("config")

		if apiKey == "" {
			fmt.Println("Error: --api-key is required")
			os.Exit(1)
		}

		cfg, err := config.Load(configPath)
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		apiClient := api.NewAPIClient(apiKey, profileID)
		if profileID == "" {
			profiles, _ := apiClient.ListProfiles()
			if len(profiles) > 0 {
				profileID = profiles[0]["id"].(string)
				apiClient.SetProfileID(profileID)
			}
		}

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		fmt.Println("Starting NextDNS Client Daemon...")
		fmt.Println("Press Ctrl+C to stop")

		// Sync disabled apps to denylist on startup
		api.SyncDisabledApps(apiClient, cfg)

		timer.StartDaemon(ctx, apiClient, cfg, configPath)

		fmt.Println("Daemon exited cleanly.")
	},
}

func Execute() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug, // Setting min level to Debug
	}
	f, err := os.OpenFile("./app.log", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		// Handle error (e.g., log to stderr or panic)
		panic(err)
	}
	handler := slog.NewJSONHandler(f, opts)
	slog.SetDefault(slog.New(handler))
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringP("api-key", "k", "", "API Key for NextDNS authentication")
	rootCmd.Flags().StringP("profile-id", "p", "", "Profile ID for NextDNS configuration")
	rootCmd.Flags().StringP("config", "c", "config.yaml", "Path to configuration file")
	rootCmd.Flags().BoolP("debug", "d", false, "Enable debug mode to display API logs in the UI")

	daemonCmd.Flags().StringP("api-key", "k", "", "API Key for NextDNS authentication")
	daemonCmd.Flags().StringP("profile-id", "p", "", "Profile ID for NextDNS configuration")
	daemonCmd.Flags().StringP("config", "c", "config.yaml", "Path to configuration file")

	rootCmd.AddCommand(daemonCmd)
}
