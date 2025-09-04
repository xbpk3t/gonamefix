package gonamefix

import (
	"go/ast"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const doc = "gonamefix checks for prohibited naming conventions and suggests replacements"

// NewAnalyzer creates a new analyzer with the given configuration
func NewAnalyzer(config Config) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "gonamefix",
		Doc:      doc,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run: func(pass *analysis.Pass) (interface{}, error) {
			return runWithConfig(pass, config)
		},
	}
}

// Analyzer is the default analyzer for gonamefix with no predefined mappings
var Analyzer = NewAnalyzer(Config{
	Check:         [][]string{}, // No default mappings
	ExcludeFiles:  []string{"*.pb.go", "*_test.go"},
	ExcludeDirs:   []string{"vendor", "node_modules", ".git"},
	CaseSensitive: false,
})


// Config represents configuration for the gonamefix linter.
type Config struct {
	// Check contains mapping of long names to short names [original, replacement]
	Check [][]string `mapstructure:"check"`
	// ExcludeFiles contains file patterns to exclude
	ExcludeFiles []string `mapstructure:"exclude-files"`
	// ExcludeDirs contains directory patterns to exclude
	ExcludeDirs []string `mapstructure:"exclude-dirs"`
	// CaseSensitive controls whether the matching is case sensitive (default: false for camelCase)
	CaseSensitive bool `mapstructure:"case-sensitive"`
}

type namePattern struct {
	regex       *regexp.Regexp
	original    string
	replacement string
}

func runWithConfig(pass *analysis.Pass, config Config) (interface{}, error) {

	// Skip if file should be excluded
	filename := pass.Fset.Position(pass.Files[0].Pos()).Filename
	if shouldExcludeFile(filename, config) {
		return nil, nil
	}

	// Build name mappings from config
	nameMappings := buildNameMappings(config.Check)
	if len(nameMappings) == 0 {
		return nil, nil
	}

	// Compile regex patterns
	patterns := buildPatterns(nameMappings, config.CaseSensitive)

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.Ident)(nil),
		(*ast.FuncDecl)(nil),
		(*ast.TypeSpec)(nil),
		(*ast.ValueSpec)(nil),
		(*ast.Field)(nil),
	}

	// Track checked identifiers to avoid duplicates
	checked := make(map[*ast.Ident]bool)

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Name != nil && !checked[node.Name] {
				checkIdentifier(pass, node.Name, patterns, config.CaseSensitive)
				checked[node.Name] = true
			}
			// Check function parameters
			if node.Type != nil && node.Type.Params != nil {
				for _, param := range node.Type.Params.List {
					for _, name := range param.Names {
						if !checked[name] {
							checkIdentifier(pass, name, patterns, config.CaseSensitive)
							checked[name] = true
						}
					}
				}
			}
			// Check function results
			if node.Type != nil && node.Type.Results != nil {
				for _, result := range node.Type.Results.List {
					for _, name := range result.Names {
						if !checked[name] {
							checkIdentifier(pass, name, patterns, config.CaseSensitive)
							checked[name] = true
						}
					}
				}
			}
		case *ast.TypeSpec:
			if node.Name != nil && !checked[node.Name] {
				checkIdentifier(pass, node.Name, patterns, config.CaseSensitive)
				checked[node.Name] = true
			}
		case *ast.ValueSpec:
			for _, name := range node.Names {
				if !checked[name] {
					checkIdentifier(pass, name, patterns, config.CaseSensitive)
					checked[name] = true
				}
			}
		case *ast.Field:
			for _, name := range node.Names {
				if !checked[name] {
					checkIdentifier(pass, name, patterns, config.CaseSensitive)
					checked[name] = true
				}
			}
		}
	})

	return nil, nil
}

func buildNameMappings(check [][]string) map[string]string {
	mappings := make(map[string]string)
	for _, pair := range check {
		if len(pair) == 2 {
			mappings[pair[0]] = pair[1]
		}
	}
	return mappings
}

func buildPatterns(mappings map[string]string, caseSensitive bool) []namePattern {
	var patterns []namePattern
	for original, replacement := range mappings {
		var regex *regexp.Regexp
		var err error

		// Create pattern that matches the word exactly or as part of camelCase
		if caseSensitive {
			// Match exact word or camelCase patterns
			pattern := `\b` + regexp.QuoteMeta(original) + `\b|` + regexp.QuoteMeta(strings.Title(original))
			regex, err = regexp.Compile(pattern)
		} else {
			// Case insensitive matching
			pattern := `(?i)\b` + regexp.QuoteMeta(original) + `\b|` + regexp.QuoteMeta(strings.Title(original))
			regex, err = regexp.Compile(pattern)
		}

		if err == nil {
			patterns = append(patterns, namePattern{
				regex:       regex,
				original:    original,
				replacement: replacement,
			})
		}
	}
	return patterns
}

func checkIdentifier(pass *analysis.Pass, ident *ast.Ident, patterns []namePattern, caseSensitive bool) {
	if ident == nil || ident.Name == "" {
		return
	}

	// Skip if it's an exact Go keyword match (only single words)
	if isGoKeyword(ident.Name) {
		return
	}

	name := ident.Name

	for _, pattern := range patterns {
		suggestedName := replaceInName(name, pattern.original, pattern.replacement, caseSensitive)

		if suggestedName != name {
			pass.Reportf(ident.Pos(), "suggest replacing '%s' with '%s'", name, suggestedName)
			break // Only report the first match to avoid duplicate reports
		}
	}
}

func replaceInName(name, original, replacement string, caseSensitive bool) string {
	if name == "" || original == "" {
		return name
	}

	// Check for exact match (case sensitive or insensitive)
	if caseSensitive && name == original {
		return replacement
	} else if !caseSensitive && strings.EqualFold(name, original) {
		// Preserve case style for case insensitive match
		if len(name) > 0 && isUpperCase(rune(name[0])) {
			return strings.Title(replacement)
		}
		return replacement
	}

	// Check for camelCase patterns
	return replaceCamelCase(name, original, replacement, caseSensitive)
}

func replaceCamelCase(name, original, replacement string, caseSensitive bool) string {
	result := name

	if !caseSensitive {
		lowerName := strings.ToLower(name)
		lowerOriginal := strings.ToLower(original)

		// Check if original is at the beginning (case insensitive)
		if strings.HasPrefix(lowerName, lowerOriginal) {
			if len(name) == len(original) {
				// Preserve original case style
				if isUpperCase(rune(name[0])) {
					return strings.Title(replacement)
				}
				return replacement
			} else if len(name) > len(original) && isUpperCase(rune(name[len(original)])) {
				// camelCase: requestHandler -> reqHandler
				if isUpperCase(rune(name[0])) {
					return strings.Title(replacement) + name[len(original):]
				}
				return replacement + name[len(original):]
			}
		}

		// Check if original is embedded in camelCase
		titleOriginal := strings.Title(original)
		if idx := strings.Index(name, titleOriginal); idx > 0 {
			// Make sure it's a proper word boundary
			if idx+len(titleOriginal) == len(name) ||
				(idx+len(titleOriginal) < len(name) && isUpperCase(rune(name[idx+len(titleOriginal)]))) {
				return name[:idx] + strings.Title(replacement) + name[idx+len(titleOriginal):]
			}
		}
	} else {
		// Case sensitive matching (similar logic but without toLower)
		if strings.HasPrefix(name, original) {
			if len(name) == len(original) {
				return replacement
			} else if len(name) > len(original) && isUpperCase(rune(name[len(original)])) {
				return replacement + name[len(original):]
			}
		}

		titleOriginal := strings.Title(original)
		if idx := strings.Index(name, titleOriginal); idx > 0 {
			if idx+len(titleOriginal) == len(name) ||
				(idx+len(titleOriginal) < len(name) && isUpperCase(rune(name[idx+len(titleOriginal)]))) {
				return name[:idx] + strings.Title(replacement) + name[idx+len(titleOriginal):]
			}
		}
	}

	return result
}

func isUpperCase(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

func isGoKeyword(name string) bool {
	// Only match exact Go keywords, not compound words
	keywords := map[string]bool{
		"break":       true,
		"case":        true,
		"chan":        true,
		"const":       true,
		"continue":    true,
		"default":     true,
		"defer":       true,
		"else":        true,
		"fallthrough": true,
		"for":         true,
		"func":        true,
		"go":          true,
		"goto":        true,
		"if":          true,
		"import":      true,
		"interface":   true,
		"map":         true,
		"package":     true,
		"range":       true,
		"return":      true,
		"select":      true,
		"struct":      true,
		"switch":      true,
		"type":        true,
		"var":         true,
	}

	return keywords[name]
}

func shouldExcludeFile(filename string, config Config) bool {
	base := filepath.Base(filename)
	for _, pattern := range config.ExcludeFiles {
		matched, err := filepath.Match(pattern, base)
		if err == nil && matched {
			return true
		}
	}

	for _, pattern := range config.ExcludeDirs {
		if strings.Contains(filename, pattern) {
			return true
		}
	}

	return false
}
