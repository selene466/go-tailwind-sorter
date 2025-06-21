package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"

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

		totalViolations, fixableViolations := processFileResults(fileResults, fix)

		if totalViolations > 0 {
			utils.PrintSummary(totalViolations, fixableViolations, fix)
			if !fix {
				os.Exit(1)
			}
		} else {
			fmt.Fprintln(os.Stderr, color.GreenString("✨ All files are sorted."))
		}

	},
}

func processFileResults(fileResults []service.FileResult, shouldFix bool) (totalViolations, fixableViolations int) {
	for _, fileResult := range fileResults {
		if fileResult.Err != nil {
			fmt.Fprintln(os.Stderr, color.RedString("Error processing %s: %w", fileResult.FilePath, fileResult.Err))
			continue
		}

		if len(fileResult.Violations) > 0 {
			for _, violation := range fileResult.Violations {
				totalViolations++
				if violation.Fixable {
					fixableViolations++
				}

				if !shouldFix {
					processViolation(fileResult.FilePath, fileResult.OriginalBytes, violation)
				}
			}
		}
	}

	return totalViolations, fixableViolations
}

func processViolation(filePath string, content []byte, violation service.Violation) {
	pathColor := color.New(color.Bold)
	ruleCodeColor := color.New(color.FgRed)
	fixMarkerColor := color.New(color.Faint)
	lineNumberColor := color.New(color.FgBlue, color.Faint)
	pipeColor := color.New(color.FgRed, color.Faint)
	pointerColor := color.New(color.FgRed)
	helpColor := color.New(color.FgBlue)

	fixMarker := ""
	if violation.Fixable {
		fixMarker = fixMarkerColor.Sprint(" [*]")
	}
	fmt.Fprintf(os.Stderr, "%s:%d:%d: %s%s %s\n", pathColor.Sprint(filePath), violation.Line, violation.Col, ruleCodeColor.Sprint(violation.Rule), fixMarker, violation.Msg)

	scanner := bufio.NewScanner(bytes.NewReader(content))

	const contextLines = 1 // TODO: Make this configurable in the future.
	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	startLine := max(violation.Line-1-contextLines, 0)

	endLine := violation.Line - 1 + contextLines
	if endLine >= len(lines) {
		endLine = len(lines) - 1
	}

	maxLineNumWidth := len(strconv.Itoa(endLine + 1))

	for idx := startLine; idx <= endLine; idx++ {
		lineNumStr := strconv.Itoa(idx + 1)
		paddedLineNum := fmt.Sprintf("%*s", maxLineNumWidth, lineNumStr)

		fmt.Fprintf(os.Stderr, "  %s %s %s\n", lineNumberColor.Sprint(paddedLineNum), pipeColor.Sprint("|"), lines[idx])

		if idx == violation.Line-1 {
			pointerWidth := max(violation.EndOffset-violation.StartOffset, 1)

			fmt.Fprintf(os.Stderr, "  %s %s %s%s %s\n", strings.Repeat(" ", maxLineNumWidth), pipeColor.Sprint("|"), strings.Repeat(" ", violation.Col-1), pointerColor.Sprint(strings.Repeat("^", pointerWidth)), pointerColor.Sprint(violation.Rule))
		}
	}

	fmt.Fprintf(os.Stderr, "  %s %s %s\n\n", helpColor.Sprint("="), color.New(color.FgCyan).Sprint("help:"), helpColor.Sprint("Sort the Tailwind CSS classes in the attribute"))
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = Version
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "Path to a custom TOML config file.")

	rootCmd.Flags().BoolVar(&fix, "fix", false, "Apply fixes to the files.")
}
