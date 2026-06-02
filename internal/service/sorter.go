package service

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/selene466/go-tailwind-sorter/internal/config"
	"github.com/selene466/go-tailwind-sorter/internal/utils"
)

var templateLiteralSplitRegex *regexp.Regexp = regexp.MustCompile(`(?s)(\$\{.+?\})`)

type Sorter struct {
	Fix     bool
	Workers int
	Config  *config.Config

	classAttributesRegex *regexp.Regexp
	classTrie            *utils.PrefixTrie
}

func SorterServiceNew(config *config.Config, fix bool, workers int) (*Sorter, error) {
	regexPattern := fmt.Sprintf(`(?s)((?:%s))(\s*=\s*)`+`(?:((["])(.*?)(["]))|((['])(.*?)([']))|(([`+"`"+`])(.*?)([`+"`"+`])))`, strings.Join(config.ClassAttributes, "|"))

	classAttributesRegex, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid classAttributes pattern: %w", err)
	}

	trie := utils.NewPrefixTrie()
	for idx, prefix := range config.ClassOrder {
		trie.Insert(prefix, idx)
	}

	return &Sorter{
		Fix:     fix,
		Workers: workers,
		Config:  config,

		classAttributesRegex: classAttributesRegex,
		classTrie:            trie,
	}, nil
}

type VariantProperty struct {
	Order int
	Name  string
}

type ClassProperty struct {
	Variants     []VariantProperty
	UtilityOrder int
	OriginalName string
}

func (sorter *Sorter) getClassProperty(className string) ClassProperty {
	parts := utils.SplitVariants(className)
	variants := make([]VariantProperty, 0, len(parts)-1)

	utilityIndex := len(parts) - 1
	for idx, part := range parts {
		// Exact Match (e.g. "hover", "sm", "dark")
		if order, ok := sorter.Config.VariantOrder[part]; ok {
			variants = append(variants, VariantProperty{Order: order, Name: part})
			continue
		}

		// Dynamic Prefix Match (e.g. "group-hover", "has-[.foo]", "@max-md")
		matched := false
		if dashIdx := strings.IndexAny(part, "-[("); dashIdx != -1 {
			basePrefix := part[:dashIdx+1] // Extracts "group-" or "has-" or "@max-"
			if order, ok := sorter.Config.VariantOrder[basePrefix]; ok {
				variants = append(variants, VariantProperty{Order: order, Name: part})
				matched = true
			}
		}

		if matched {
			continue
		}

		// Complete Arbitrary Variants (e.g. "[&_p]", "@min-(600px)")
		if strings.HasPrefix(part, "[") || strings.HasPrefix(part, "@") {
			variants = append(variants, VariantProperty{Order: 999, Name: part})
			continue
		}

		utilityIndex = idx
		break
	}
	utility := strings.Join(parts[utilityIndex:], ":")

	sort.Slice(variants, func(i, j int) bool {
		if variants[i].Order != variants[j].Order {
			return variants[i].Order < variants[j].Order
		}
		return variants[i].Name < variants[j].Name
	})

	utilityOrder := sorter.classTrie.GetLongestPrefixOrder(utility, len(sorter.Config.ClassOrder))

	return ClassProperty{Variants: variants, UtilityOrder: utilityOrder, OriginalName: className}
}

func (sorter *Sorter) tokenizeTWClassString(twClassString string) []string {
	var tokens []string
	var currentToken strings.Builder

	bracketLevel := 0

	for _, char := range twClassString {
		switch char {
		case '[':
			bracketLevel++
			currentToken.WriteRune(char)
		case ']':
			bracketLevel--
			currentToken.WriteRune(char)
		case ' ', '\t', '\n', '\r':
			if bracketLevel == 0 {
				if currentToken.Len() > 0 {
					tokens = append(tokens, currentToken.String())
					currentToken.Reset()
				}
			} else {
				currentToken.WriteRune(char)
			}

		default:
			currentToken.WriteRune(char)
		}
	}

	if currentToken.Len() > 0 {
		tokens = append(tokens, currentToken.String())
	}

	return tokens
}

func (sorter *Sorter) sortStaticTWClassString(staticTWClassString string) string {
	fields := sorter.tokenizeTWClassString(staticTWClassString)
	if len(fields) == 0 {
		return ""
	}

	seenTWClass := make(map[string]struct{})
	uniqueTWClasses := make([]string, 0, len(fields))

	for _, twClass := range fields {
		if _, exists := seenTWClass[twClass]; !exists {
			seenTWClass[twClass] = struct{}{}
			uniqueTWClasses = append(uniqueTWClasses, twClass)
		}
	}

	sort.SliceStable(uniqueTWClasses, func(i, j int) bool {
		classIProperty, classJProperty := sorter.getClassProperty(uniqueTWClasses[i]), sorter.getClassProperty(uniqueTWClasses[j])

		if len(classIProperty.Variants) != len(classJProperty.Variants) {
			return len(classIProperty.Variants) < len(classJProperty.Variants)
		}

		for idx := 0; idx < len(classIProperty.Variants); idx++ {
			if classIProperty.Variants[idx].Order != classJProperty.Variants[idx].Order {
				return classIProperty.Variants[idx].Order < classJProperty.Variants[idx].Order
			}
		}

		return classIProperty.UtilityOrder < classJProperty.UtilityOrder
	})

	return strings.Join(uniqueTWClasses, " ")
}

func (sorter *Sorter) sortTWClassString(twClassString string) string {
	if strings.Contains(twClassString, "${") {
		parts := templateLiteralSplitRegex.Split(twClassString, -1)

		var result strings.Builder

		for idx, part := range parts {
			var processedPart string

			// Even-indexed parts are the static text between the dynamic blocks.
			// Odd-indexed parts are the dynamic blocks themselves.
			if idx%2 == 0 {
				processedPart = sorter.sortStaticTWClassString(part)
			} else {
				processedPart = part
			}

			if processedPart != "" {
				if result.Len() > 0 {
					result.WriteString(" ")
				}
				result.WriteString(processedPart)
			}
		}

		return result.String()

	} else {
		return sorter.sortStaticTWClassString(twClassString)
	}
}

func (sorter *Sorter) fileHasValidExtension(filePath string) bool {
	fileExtension := filepath.Ext(filePath)
	return slices.Contains(sorter.Config.FilePatterns, fileExtension)
}

func (sorter *Sorter) findFiles(paths []string) ([]string, error) {
	files := make(map[string]struct{})

	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("invalid path %s: %w", path, err)
		}

		if !info.IsDir() {
			if sorter.fileHasValidExtension(path) {
				files[path] = struct{}{}
			}
			continue
		}

		err = filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && sorter.fileHasValidExtension(path) {
				files[path] = struct{}{}
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to walk directory %s: %w", path, err)
		}
	}

	result := make([]string, 0, len(files))
	for file := range files {
		result = append(result, file)
	}

	sort.Strings(result)

	return result, nil
}

type Violation struct {
	Line        int
	Col         int
	StartOffset int
	EndOffset   int
	Rule        string
	Msg         string
	Fixable     bool
}

func (sorter *Sorter) analyzeAndFixContent(content []byte) ([]Violation, []byte) {
	var violations []Violation
	var result bytes.Buffer

	matches := sorter.classAttributesRegex.FindAllSubmatchIndex(content, -1)
	if len(matches) == 0 {
		return nil, content
	}

	result.Grow(len(content))
	lastIdx := 0

	for _, match := range matches {
		var startOffset, endOffset int

		if match[10] != -1 { // " "
			startOffset, endOffset = match[10], match[11]
		} else if match[18] != -1 { // ' '
			startOffset, endOffset = match[18], match[19]
		} else if match[26] != -1 { // ` `
			startOffset, endOffset = match[26], match[27]
		} else {
			continue
		}

		result.Write(content[lastIdx:startOffset])

		twClassString := string(content[startOffset:endOffset])
		sortedTWClassString := sorter.sortTWClassString(twClassString)

		result.WriteString(sortedTWClassString)
		lastIdx = endOffset

		if twClassString != sortedTWClassString {
			line, col := utils.OffsetToLineCol(content, startOffset)
			violations = append(violations, Violation{
				Line:        line,
				Col:         col,
				StartOffset: startOffset,
				EndOffset:   endOffset,
				Rule:        "TWS001",
				Msg:         "Unsorted Tailwind classes",
				Fixable:     true,
			})
		}
	}

	result.Write(content[lastIdx:])

	return violations, result.Bytes()
}

type FileResult struct {
	FilePath      string
	Violations    []Violation
	SortedBytes   []byte
	OriginalBytes []byte
	Err           error
}

func (sorter *Sorter) worker(wg *sync.WaitGroup, jobs <-chan string, results chan<- FileResult) {
	defer wg.Done()

	for filePath := range jobs {
		originalContent, err := os.ReadFile(filePath)
		if err != nil {
			results <- FileResult{Err: fmt.Errorf("reading file %s: %w", filePath, err)}
			continue
		}

		violations, sortedContent := sorter.analyzeAndFixContent(originalContent)
		if len(violations) == 0 {
			continue
		}

		results <- FileResult{
			FilePath:      filePath,
			Violations:    violations,
			SortedBytes:   sortedContent,
			OriginalBytes: originalContent,
		}
	}
}

func (sorter *Sorter) Run(paths []string) ([]FileResult, error) {
	filesToProcess, err := sorter.findFiles(paths)
	if err != nil {
		return nil, fmt.Errorf("failed to find files: %w", err)
	}

	var wg sync.WaitGroup
	jobs := make(chan string, sorter.Workers)
	results := make(chan FileResult, sorter.Workers)

	for range sorter.Workers {
		wg.Add(1)
		go sorter.worker(&wg, jobs, results)
	}

	go func() {
		for _, file := range filesToProcess {
			jobs <- file
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var fileResults []FileResult
	for result := range results {
		if result.Err != nil {
			fileResults = append(fileResults, result)
			continue
		}

		if len(result.Violations) > 0 {
			fileResults = append(fileResults, result)

			if sorter.Fix {
				info, statErr := os.Stat(result.FilePath)

				if statErr == nil {
					if err := os.WriteFile(result.FilePath, result.SortedBytes, info.Mode()); err != nil {
						result.Err = fmt.Errorf("failed to write fixes to %s: %w", result.FilePath, err)
					}
				} else {
					result.Err = fmt.Errorf("failed to read permissions for %s before writing: %w", result.FilePath, statErr)
				}
			}
		}
	}

	sort.Slice(fileResults, func(i, j int) bool {
		return fileResults[i].FilePath < fileResults[j].FilePath
	})

	return fileResults, nil
}
