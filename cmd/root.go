package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"nextdns_client/internal/api"
	"nextdns_client/internal/config"
	"nextdns_client/internal/daemon"
	"nextdns_client/internal/timer"
	"nextdns_client/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kardianos/service"
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

		// Resolve app dir for state/logging
		appDir, _ := os.Getwd()

		// Set timer state directory
		timer.SetStateDir(appDir)

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
	Long:  "Run the NextDNS daemon, either interactively or as a managed system service",
}

var daemonRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the daemon interactively",
	Run: func(cmd *cobra.Command, args []string) {
		apiKey, _ := cmd.Flags().GetString("api-key")
		profileID, _ := cmd.Flags().GetString("profile-id")
		configPath, _ := cmd.Flags().GetString("config")

		if apiKey == "" {
			fmt.Println("Error: --api-key is required")
			os.Exit(1)
		}

		program, err := daemon.NewProgram(apiKey, profileID, configPath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		svcConfig := daemon.Config()
		if service.Interactive() {
			fmt.Println("Starting NextDNS Client Daemon...")
		}

		svc, err := service.New(program, svcConfig)
		if err != nil {
			fmt.Printf("Error creating service: %v\n", err)
			os.Exit(1)
		}

		if service.Interactive() {
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
			defer signal.Stop(quit)

			if err := svc.Run(); err != nil {
				fmt.Printf("Error starting daemon: %v\n", err)
				os.Exit(1)
			}

			<-quit
			fmt.Println("\nShutting down daemon...")
		} else {
			if err := svc.Run(); err != nil {
				fmt.Printf("Error starting daemon: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

var daemonInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the daemon as a system service",
	Run: func(cmd *cobra.Command, args []string) {
		apiKey, _ := cmd.Flags().GetString("api-key")
		profileID, _ := cmd.Flags().GetString("profile-id")
		configPath, _ := cmd.Flags().GetString("config")

		program, err := daemon.NewProgram(apiKey, profileID, configPath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if err := daemon.InstallProgram(program); err != nil {
			fmt.Printf("Error installing service: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Daemon successfully installed.")
	},
}

var daemonStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the installed daemon",
	Run: func(cmd *cobra.Command, args []string) {
		program, err := daemon.NewProgram("", "", "")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if err := daemon.StartProgram(program); err != nil {
			fmt.Printf("Error starting service: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Daemon started.")
	},
}

var daemonStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the installed daemon",
	Run: func(cmd *cobra.Command, args []string) {
		program, err := daemon.NewProgram("", "", "")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if err := daemon.StopProgram(program); err != nil {
			fmt.Printf("Error stopping service: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Daemon stopped.")
	},
}

var daemonUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall the daemon",
	Run: func(cmd *cobra.Command, args []string) {
		program, err := daemon.NewProgram("", "", "")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if err := daemon.UninstallProgram(program); err != nil {
			fmt.Printf("Error uninstalling service: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Daemon uninstalled.")
	},
}

var daemonStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display the daemon status",
	Run: func(cmd *cobra.Command, args []string) {
		program, err := daemon.NewProgram("", "", "")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		status, err := daemon.GetStatusProgram(program)
		if err != nil {
			fmt.Printf("Error getting service status: %v\n", err)
			os.Exit(1)
		}

		var statusStr string
		switch status {
		case service.StatusRunning:
			statusStr = "running"
		case service.StatusStopped:
			statusStr = "stopped"
		default:
			statusStr = "unknown"
		}

		fmt.Printf("Current status: %s\n", statusStr)
	},
}

func Execute() {
	// Resolve log file to absolute path
	absLogPath, err := filepath.Abs("./app.log")
	if err != nil {
		panic(fmt.Sprintf("resolving log path: %v", err))
	}
	if err := daemon.InitCLILogger(absLogPath); err != nil {
		panic(err)
	}
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
	daemonRunCmd.Flags().StringP("api-key", "k", "", "API Key for NextDNS authentication")
	daemonRunCmd.Flags().StringP("profile-id", "p", "", "Profile ID for NextDNS configuration")
	daemonRunCmd.Flags().StringP("config", "c", "config.yaml", "Path to configuration file")

	rootCmd.AddCommand(daemonCmd)
	daemonCmd.AddCommand(daemonRunCmd)
	daemonCmd.AddCommand(daemonInstallCmd)
	daemonCmd.AddCommand(daemonStartCmd)
	daemonCmd.AddCommand(daemonStopCmd)
	daemonCmd.AddCommand(daemonUninstallCmd)
	daemonCmd.AddCommand(daemonStatusCmd)

	rootCmd.AddCommand(serviceCmd)
}
