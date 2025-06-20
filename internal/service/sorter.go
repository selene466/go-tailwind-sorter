package service

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/dexter2389/go-tailwind-sorter/internal/config"
	"github.com/fatih/color"
)

const numWorkers int = 4

var arbitraryVariantRegex *regexp.Regexp = regexp.MustCompile(`^\[.+?\]`)
var templateLiteralSplitRegex *regexp.Regexp = regexp.MustCompile(`(\$\{[^}]*\})`)

type Sorter struct {
	Check   bool
	Verbose bool
	Config  *config.Config

	classAttributesRegex *regexp.Regexp
	stderr               func(format string, a ...any)
}

func SorterServiceNew(config *config.Config, check, verbose bool) (*Sorter, error) {
	regexPattern := fmt.Sprintf(`((?:%s))(\s*=\s*)(["'`+"`"+`])(.*?)(["'`+"`"+`])`, strings.Join(config.ClassAttributes, "|"))

	classAttributesRegex, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid classAttributes pattern: %w", err)
	}

	return &Sorter{
		Check:   check,
		Verbose: verbose,
		Config:  config,

		classAttributesRegex: classAttributesRegex,
		stderr: func(format string, a ...any) {
			color.New(color.FgWhite).Fprintf(os.Stderr, format, a...)
		},
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
	parts := strings.Split(className, ":")
	variants := make([]VariantProperty, 0) // Create an empty slice

	utilityIndex := len(parts) - 1
	for idx, part := range parts {
		if arbitraryVariantRegex.MatchString(part) {
			variants = append(variants, VariantProperty{Order: 99, Name: part})
			continue
		}
		if order, ok := sorter.Config.VariantOrder[part]; ok {
			variants = append(variants, VariantProperty{Order: order, Name: part})
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

	utilityOrder := len(sorter.Config.ClassOrder)
	for idx, prefix := range sorter.Config.ClassOrder {
		if strings.HasPrefix(utility, prefix) {
			utilityOrder = idx
			break
		}
	}

	return ClassProperty{Variants: variants, UtilityOrder: utilityOrder, OriginalName: className}
}

func (sorter *Sorter) sortTWClassString(twClassString string) string {
	if strings.Contains(twClassString, "${") {
		parts := templateLiteralSplitRegex.Split(twClassString, -1)
		delimiters := templateLiteralSplitRegex.FindAllString(twClassString, -1)

		var result strings.Builder

		for idx, part := range parts {
			result.WriteString(sorter.sortTWClassString(part))
			if idx < len(delimiters) {
				result.WriteString(delimiters[idx])
			}
		}
	}

	fields := strings.Fields(twClassString)
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

func (sorter *Sorter) processFileContent(content []byte) []byte {
	return sorter.classAttributesRegex.ReplaceAllFunc(content, func(match []byte) []byte {
		parts := sorter.classAttributesRegex.FindSubmatch(match)
		return fmt.Appendf(nil, `%s%s%s%s%s`, parts[1], parts[2], parts[3], sorter.sortTWClassString(string(parts[4])), parts[5])
	})
}

func (sorter *Sorter) fileHasValidExtension(filePath string) bool {
	fileExtension := filepath.Ext(filePath)

	for _, pattern := range sorter.Config.FilePatterns {
		if fileExtension == pattern {
			return true
		}
	}

	return false
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

func (sorter *Sorter) printSummary(filesToFormat []string) error {
	if sorter.Check {
		if len(filesToFormat) > 0 {
			sorter.stderr(color.RedString("\nError: The following files are not formatted correctly:\n"))
			for _, file := range filesToFormat {
				sorter.stderr(color.YellowString("- %s\n", file))
			}
			return fmt.Errorf("%d files need formatting", len(filesToFormat))
		}
		sorter.stderr(color.GreenString("\n✅ All files are formatted correctly.\n"))
	} else if len(filesToFormat) > 0 {
		sorter.stderr(color.GreenString("\n✅ Successfully formatted %d file(s).\n", len(filesToFormat)))
	} else {
		sorter.stderr(color.GreenString("\n✨ All files were already formatted.\n"))
	}
	return nil
}

type processResult struct {
	filePath        string
	changed         bool
	originalContent string
	sortedContent   string
	err             error
}

func (sorter *Sorter) worker(wg *sync.WaitGroup, jobs <-chan string, results chan<- processResult) {
	defer wg.Done()

	for filePath := range jobs {
		if sorter.Verbose {
			sorter.stderr("Processing %s...\n", color.CyanString(filePath))
		}

		originalFileContent, err := os.ReadFile(filePath)
		if err != nil {
			results <- processResult{err: fmt.Errorf("reading file %s: %w", filePath, err)}
			continue
		}

		sortedFileContent := sorter.processFileContent(originalFileContent)
		results <- processResult{
			filePath:        filePath,
			changed:         !bytes.Equal(originalFileContent, sortedFileContent),
			originalContent: string(originalFileContent),
			sortedContent:   string(sortedFileContent),
		}
	}
}

func (sorter *Sorter) Run(paths []string) error {
	filesToProcess, err := sorter.findFiles(paths)

	if err != nil {
		return fmt.Errorf("failed to find files: %w", err)
	}

	if len(filesToProcess) == 0 {
		sorter.stderr(color.YellowString("Warning: No files found to process\n"))
		return nil
	}

	var wg sync.WaitGroup
	jobs := make(chan string, len(filesToProcess))
	results := make(chan processResult, len(filesToProcess))

	for idx := 0; idx < numWorkers; idx++ {
		wg.Add(1)
		go sorter.worker(&wg, jobs, results)
	}

	for _, file := range filesToProcess {
		jobs <- file
	}

	close(jobs)
	wg.Wait()
	close(results)

	var filesToFormat []string
	for result := range results {
		if result.err != nil {
			sorter.stderr(color.RedString("Error: %v\n", result.err))
			continue
		}
		if result.changed {
			filesToFormat = append(filesToFormat, result.filePath)

			if !sorter.Check {
				sorter.stderr("Formatting %s\n", color.YellowString(result.filePath))
				if err := os.WriteFile(result.filePath, []byte(result.sortedContent), 0644); err != nil {
					sorter.stderr(color.RedString("Error writing file %s: %v\n", result.filePath, err))
				}
			}
		}
	}

	return sorter.printSummary(filesToFormat)
}
