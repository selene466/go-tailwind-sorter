package utils

import (
	"fmt"
	"os"
	"strings"

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

func SplitVariants(className string) []string {
	var parts []string
	var current strings.Builder
	bracketLevel := 0
	parenLevel := 0

	for _, char := range className {
		switch char {
		case '[':
			bracketLevel++
		case ']':
			bracketLevel--
		case '(':
			parenLevel++
		case ')':
			parenLevel--
		case ':':
			if bracketLevel == 0 && parenLevel == 0 {
				parts = append(parts, current.String())
				current.Reset()
				continue
			}
		}
		current.WriteRune(char)
	}
	parts = append(parts, current.String())
	return parts
}

type TrieNode struct {
	children map[rune]*TrieNode
	order    int
}

func NewTrieNode() *TrieNode {
	return &TrieNode{
		children: make(map[rune]*TrieNode),
		order:    -1,
	}
}

type PrefixTrie struct {
	root *TrieNode
}

func NewPrefixTrie() *PrefixTrie {
	return &PrefixTrie{root: NewTrieNode()}
}

func (trie *PrefixTrie) Insert(prefix string, order int) {
	node := trie.root
	for _, ch := range prefix {
		if _, exists := node.children[ch]; !exists {
			node.children[ch] = NewTrieNode()
		}
		node = node.children[ch]
	}
	node.order = order
}

func (trie *PrefixTrie) GetLongestPrefixOrder(s string, defaultOrder int) int {
	node := trie.root
	bestOrder := defaultOrder

	for _, ch := range s {
		if nextNode, exists := node.children[ch]; exists {
			node = nextNode
			if node.order != -1 {
				bestOrder = node.order
			}
		} else {
			break
		}
	}
	return bestOrder
}
