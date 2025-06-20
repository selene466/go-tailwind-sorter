package service

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/dexter2389/go-tailwind-sorter/internal/config"
	"github.com/dexter2389/go-tailwind-sorter/internal/utils"
)

const numWorkers int = 4

var arbitraryVariantRegex *regexp.Regexp = regexp.MustCompile(`^\[.+?\]`)
var templateLiteralSplitRegex *regexp.Regexp = regexp.MustCompile(`(\$\{[^}]*\})`)

type Sorter struct {
	Fix    bool
	Config *config.Config

	classAttributesRegex *regexp.Regexp
}

func SorterServiceNew(config *config.Config, fix bool) (*Sorter, error) {
	regexPattern := fmt.Sprintf(`((?:%s))(\s*=\s*)(["'`+"`"+`])(.*?)(["'`+"`"+`])`, strings.Join(config.ClassAttributes, "|"))

	classAttributesRegex, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid classAttributes pattern: %w", err)
	}

	return &Sorter{
		Fix:    fix,
		Config: config,

		classAttributesRegex: classAttributesRegex,
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

type Violation struct {
	Line int
	Col  int
	Rule string
	Msg  string
}

func (sorter *Sorter) findViolations(content []byte) []Violation {
	var violations []Violation

	matches := sorter.classAttributesRegex.FindAllSubmatchIndex(content, -1)
	for _, match := range matches {
		// match[8] and match[9] are the start and end of the tw_class string itself.
		startOffset, endOffset := match[8], match[9]

		twClassString := string(content[startOffset:endOffset])
		sortedTWClassString := sorter.sortTWClassString(twClassString)

		if twClassString != sortedTWClassString {
			line, col := utils.OffsetToLineCol(content, startOffset)
			violations = append(violations, Violation{
				Line: line,
				Col:  col,
				Rule: "TWS001",
				Msg:  "Unsorted Tailwind classes",
			})
		}

	}

	return violations
}

type FileResult struct {
	FilePath    string
	Violations  []Violation
	SortedBytes []byte
	Err         error
}

func (sorter *Sorter) worker(wg *sync.WaitGroup, jobs <-chan string, results chan<- FileResult) {
	defer wg.Done()

	for filePath := range jobs {
		originalContent, err := os.ReadFile(filePath)
		if err != nil {
			results <- FileResult{Err: fmt.Errorf("reading file %s: %w", filePath, err)}
			continue
		}

		violations := sorter.findViolations(originalContent)
		if len(violations) == 0 {
			continue
		}

		sortedContent := sorter.processFileContent(originalContent)
		results <- FileResult{
			FilePath:    filePath,
			Violations:  violations,
			SortedBytes: sortedContent,
		}
	}
}

func (sorter *Sorter) Run(paths []string) ([]FileResult, error) {
	filesToProcess, err := sorter.findFiles(paths)
	if err != nil {
		return nil, fmt.Errorf("failed to find files: %w", err)
	}

	var wg sync.WaitGroup
	jobs := make(chan string, len(filesToProcess))
	results := make(chan FileResult, len(filesToProcess))

	for range numWorkers {
		wg.Add(1)
		go sorter.worker(&wg, jobs, results)
	}

	for _, file := range filesToProcess {
		jobs <- file
	}

	close(jobs)
	wg.Wait()
	close(results)

	var fileResults []FileResult
	for result := range results {
		if result.Err != nil {
			fileResults = append(fileResults, result)
			continue
		}

		if len(result.Violations) > 0 {
			fileResults = append(fileResults, result)

			if sorter.Fix {
				if err := os.WriteFile(result.FilePath, result.SortedBytes, 0644); err != nil {
					result.Err = fmt.Errorf("failed to write fixes to %s: %w", result.FilePath, err)
				}
			}
		}
	}

	sort.Slice(fileResults, func(i, j int) bool {
		return fileResults[i].FilePath < fileResults[j].FilePath
	})

	return fileResults, nil
}
