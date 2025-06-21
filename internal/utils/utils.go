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

func PrintSummary(totalViolations, fixableViolations int, shouldFix bool) {
	var violationStr string
	if totalViolations == 1 {
		violationStr = "violation"
	} else {
		violationStr = "violations"
	}

	if shouldFix {
		fmt.Fprintln(os.Stderr, color.GreenString("Found %d %s (%d fixed, %d remaining).", totalViolations, violationStr, fixableViolations, totalViolations-fixableViolations))
	} else {
		fmt.Fprintln(os.Stderr, color.RedString("Found %d %s.", totalViolations, violationStr))
		if fixableViolations > 0 {
			fmt.Fprintf(os.Stderr, "%s %d potentially fixable with the --fix option.\n", color.New(color.Bold, color.Faint).Sprint("[*]"), fixableViolations)
		}
	}
}
