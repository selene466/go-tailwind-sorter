package utils

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

func OffsetToLineCol(content []byte, offset int) (line, col int) {
	line = 1
	lastNewline := -1

	if offset > len(content) {
		offset = len(content)
	}

	for idx, char := range content[:offset] {
		if char == '\n' {
			line++
			lastNewline = idx
		}
	}

	col = offset - lastNewline

	return line, col
}

func PrintSummary(totalViolations int, fixed bool) {
	if fixed {
		fmt.Fprintln(os.Stderr, color.GreenString("Found %d violation(s) and fixed them.", totalViolations))
	} else {
		fmt.Fprintln(os.Stderr, color.RedString("Found %d violation(s).", totalViolations))
		fmt.Fprintln(os.Stderr, color.New(color.Faint).Sprint("Run with --fix to apply changes."))
	}
}
