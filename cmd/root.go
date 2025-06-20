package cmd

import (
	"fmt"
	"os"

	"github.com/dexter2389/go-tailwind-sorter/internal/config"
	"github.com/dexter2389/go-tailwind-sorter/internal/service"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var check bool
var verbose bool
var configFile string
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:   "tailwind-sorter [PATH...]",
	Short: "A Go-based sorter for Tailwind CSS classes.",
	Long:  "A fast, standalone binary to sort Tailwind CSS classes in your project files.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		config, err := config.New(configFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, color.RedString("Error loading configuration: %v", err))
			os.Exit(1)
		}

		sorterService, err := service.SorterServiceNew(config, check, verbose)
		if err != nil {
			fmt.Fprintln(os.Stderr, color.RedString("Error initializing sorter: %v", err))
			os.Exit(1)
		}

		if err := sorterService.Run(args); err != nil {
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = Version
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Path to a custom TOML config file.")

	rootCmd.Flags().BoolVar(&check, "check", false, "Check if files are sorted without making changes.")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output.")
}
