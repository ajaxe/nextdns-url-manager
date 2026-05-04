package main

import (
	"fmt"
	"os"
	"path/filepath"

	"nextdns_client/internal/daemon"

	"github.com/kardianos/service"
	"github.com/spf13/cobra"
)

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manage the application as an OS-native service",
	Long: `Manage the NextDNS Client as a background service running under the OS service manager.

Available subcommands:
  install   Register the service with the OS
  uninstall Remove the service registration
  start     Start the registered service
  stop      Stop the registered service
  run       Entry point for the OS service manager (internal)`,
}

var serviceInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Register the service with the OS",
	Run: func(cmd *cobra.Command, args []string) {
		apiKey, _ := cmd.Flags().GetString("api-key")
		profileID, _ := cmd.Flags().GetString("profile-id")
		configPath, _ := cmd.Flags().GetString("config")
		displayName, _ := cmd.Flags().GetString("display-name")

		if apiKey == "" {
			fmt.Println("Error: --api-key is required for service installation")
			os.Exit(1)
		}

		// Resolve to absolute path
		if abs, err := filepath.Abs(configPath); err == nil {
			configPath = abs
		}

		program, err := daemon.NewProgram(apiKey, profileID, configPath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		svcConfig := daemon.Config()
		svcConfig.Arguments = program.GetServiceConfigArgs()

		if displayName != "" {
			svcConfig.DisplayName = displayName
		}

		if err := daemon.Install(svcConfig); err != nil {
			fmt.Printf("Error installing service: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Service successfully installed.")
	},
}

var serviceRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Entry point for the OS service manager (internal)",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		apiKey, _ := cmd.Flags().GetString("api-key")
		profileID, _ := cmd.Flags().GetString("profile-id")
		configPath, _ := cmd.Flags().GetString("config")

		program, err := daemon.NewProgram(apiKey, profileID, configPath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		svcConfig := daemon.Config()
		svcConfig.Arguments = program.GetServiceConfigArgs()

		svc, err := service.New(program, svcConfig)
		if err != nil {
			fmt.Printf("Error creating service: %v\n", err)
			os.Exit(1)
		}

		if err := svc.Run(); err != nil {
			fmt.Printf("Error running service: %v\n", err)
			os.Exit(1)
		}
	},
}

var serviceStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the registered service",
	Run: func(cmd *cobra.Command, args []string) {
		svcConfig := daemon.Config()

		if err := daemon.StartService(svcConfig); err != nil {
			fmt.Printf("Error starting service: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Service started.")
	},
}

var serviceStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the registered service",
	Run: func(cmd *cobra.Command, args []string) {
		svcConfig := daemon.Config()

		if err := daemon.StopService(svcConfig); err != nil {
			fmt.Printf("Error stopping service: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Service stopped.")
	},
}

var serviceUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove the service registration",
	Run: func(cmd *cobra.Command, args []string) {
		svcConfig := daemon.Config()

		if err := daemon.UninstallService(svcConfig); err != nil {
			fmt.Printf("Error uninstalling service: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Service uninstalled.")
	},
}

var serviceStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display the service status",
	Run: func(cmd *cobra.Command, args []string) {
		svcConfig := daemon.Config()

		status, err := daemon.GetServiceStatus(svcConfig)
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

func init() {
	serviceInstallCmd.Flags().StringP("api-key", "k", "", "API Key for NextDNS authentication (required)")
	serviceInstallCmd.Flags().StringP("profile-id", "p", "", "Profile ID for NextDNS configuration")
	serviceInstallCmd.Flags().StringP("config", "c", "config.yaml", "Path to configuration file")
	serviceInstallCmd.Flags().String("display-name", "", "Display name for the service")

	serviceRunCmd.Flags().StringP("api-key", "k", "", "API Key for NextDNS authentication")
	serviceRunCmd.Flags().StringP("profile-id", "p", "", "Profile ID for NextDNS configuration")
	serviceRunCmd.Flags().StringP("config", "c", "config.yaml", "Path to configuration file")

	serviceCmd.AddCommand(serviceInstallCmd)
	serviceCmd.AddCommand(serviceStartCmd)
	serviceCmd.AddCommand(serviceStopCmd)
	serviceCmd.AddCommand(serviceUninstallCmd)
	serviceCmd.AddCommand(serviceRunCmd)
	serviceCmd.AddCommand(serviceStatusCmd)
}
