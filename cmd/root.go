package cmd

import (
	"fmt"
	"os"

	"github.com/dexter2389/go-tailwind-sorter/internal/config"
	"github.com/dexter2389/go-tailwind-sorter/internal/service"
	"github.com/dexter2389/go-tailwind-sorter/internal/utils"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var fix bool
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

		sorterService, err := service.SorterServiceNew(config, fix)
		if err != nil {
			fmt.Fprintln(os.Stderr, color.RedString("Error initializing sorter: %v", err))
			os.Exit(1)
		}

		fileResults, err := sorterService.Run(args)
		if err != nil {
			fmt.Fprintln(os.Stderr, color.RedString("Error during execution: %v", err))
			os.Exit(1)
		}

		totalViolations := processFileResultsToGetViolations(fileResults)

		if totalViolations > 0 {
			utils.PrintSummary(totalViolations, fix)
			os.Exit(1)
		} else {
			fmt.Fprintln(os.Stderr, color.GreenString("✨ All files are sorted."))
		}

	},
}

func processFileResultsToGetViolations(fileResults []service.FileResult) int {
	totalViolations := 0

	for _, fileResult := range fileResults {
		if fileResult.Err != nil {
			fmt.Fprintln(os.Stderr, color.RedString("Error processing %s: %w", fileResult.FilePath, fileResult.Err))
			continue
		}

		if len(fileResult.Violations) > 0 {
			fmt.Fprintln(os.Stderr, color.New(color.Bold).Sprint(fileResult.FilePath))
			for _, violation := range fileResult.Violations {
				lineCol := fmt.Sprintf("%d:%d", violation.Line, violation.Col)
				fmt.Fprintf(os.Stderr, "  %s  %s  %s\n", color.New(color.Faint).Sprint(lineCol), color.RedString(violation.Rule), violation.Msg)
				totalViolations++
			}
		}
	}

	return totalViolations
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = Version
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Path to a custom TOML config file.")

	rootCmd.Flags().BoolVar(&fix, "fix", false, "Apply fixes to the files.")
}
