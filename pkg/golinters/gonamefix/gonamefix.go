package gonamefix

import (
	"fmt"
	"go/ast"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const LinterName = "gonamefix"

// Config represents configuration for the gonamefix linter.
type Config struct {
	// Check contains mapping of long names to short names [original, replacement]
	Check [][]string `mapstructure:"check"`
	// ExcludeFiles contains file patterns to exclude
	ExcludeFiles []string `mapstructure:"exclude-files"`
	// ExcludeDirs contains directory patterns to exclude
	ExcludeDirs []string `mapstructure:"exclude-dirs"`
	// CaseSensitive controls whether the matching is case sensitive
	CaseSensitive bool `mapstructure:"case-sensitive"`
}

// NewAnalyzer returns a new analyzer for gonamefix.
func NewAnalyzer(config Config) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     LinterName,
		Doc:      "gonamefix checks for prohibited naming conventions and suggests replacements",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run: func(pass *analysis.Pass) (interface{}, error) {
			return run(pass, config)
		},
	}
}

func run(pass *analysis.Pass, config Config) (interface{}, error) {
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

	// Compile regex patterns for case-insensitive matching if needed
	patterns := buildPatterns(nameMappings, config.CaseSensitive)

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.Ident)(nil),
		(*ast.FuncDecl)(nil),
		(*ast.TypeSpec)(nil),
		(*ast.ValueSpec)(nil),
		(*ast.Field)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		switch node := n.(type) {
		case *ast.Ident:
			checkIdentifier(pass, node, patterns, config.CaseSensitive)
		case *ast.FuncDecl:
			if node.Name != nil {
				checkIdentifier(pass, node.Name, patterns, config.CaseSensitive)
			}
			// Check function parameters
			if node.Type != nil && node.Type.Params != nil {
				for _, param := range node.Type.Params.List {
					for _, name := range param.Names {
						checkIdentifier(pass, name, patterns, config.CaseSensitive)
					}
				}
			}
			// Check function results
			if node.Type != nil && node.Type.Results != nil {
				for _, result := range node.Type.Results.List {
					for _, name := range result.Names {
						checkIdentifier(pass, name, patterns, config.CaseSensitive)
					}
				}
			}
		case *ast.TypeSpec:
			if node.Name != nil {
				checkIdentifier(pass, node.Name, patterns, config.CaseSensitive)
			}
		case *ast.ValueSpec:
			for _, name := range node.Names {
				checkIdentifier(pass, name, patterns, config.CaseSensitive)
			}
		case *ast.Field:
			for _, name := range node.Names {
				checkIdentifier(pass, name, patterns, config.CaseSensitive)
			}
		}
	})

	return nil, nil
}

type namePattern struct {
	regex       *regexp.Regexp
	original    string
	replacement string
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

		if caseSensitive {
			regex, err = regexp.Compile(fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(original)))
		} else {
			regex, err = regexp.Compile(fmt.Sprintf(`(?i)\b%s\b`, regexp.QuoteMeta(original)))
		}

		if err == nil {
			patterns = append(patterns, namePattern{
				original:    original,
				replacement: replacement,
				regex:       regex,
			})
		}
	}

	return patterns
}

func checkIdentifier(pass *analysis.Pass, ident *ast.Ident, patterns []namePattern, caseSensitive bool) {
	if ident == nil || ident.Name == "" {
		return
	}

	// Skip built-in types and common Go identifiers
	if isBuiltinIdentifier(ident.Name) {
		return
	}

	name := ident.Name

	for _, pattern := range patterns {
		var matches bool
		var suggestedName string

		if caseSensitive {
			matches = pattern.regex.MatchString(name)
		} else {
			matches = pattern.regex.MatchString(strings.ToLower(name))
		}

		if matches {
			if caseSensitive {
				suggestedName = pattern.regex.ReplaceAllString(name, pattern.replacement)
			} else {
				// For case-insensitive matching, we need to preserve the case of non-matching parts
				suggestedName = replacePreservingCase(name, pattern.original, pattern.replacement)
			}

			if suggestedName != name {
				diagnostic := analysis.Diagnostic{
					Pos:     ident.Pos(),
					End:     ident.End(),
					Message: fmt.Sprintf("Replace '%s' with '%s'", name, suggestedName),
					SuggestedFixes: []analysis.SuggestedFix{
						{
							Message: fmt.Sprintf("Replace '%s' with '%s'", name, suggestedName),
							TextEdits: []analysis.TextEdit{
								{
									Pos:     ident.Pos(),
									End:     ident.End(),
									NewText: []byte(suggestedName),
								},
							},
						},
					},
				}
				pass.Report(diagnostic)

				break // Only report the first match to avoid duplicate reports
			}
		}
	}
}

func replacePreservingCase(input, original, replacement string) string {
	// Handle camelCase preservation
	if strings.Contains(strings.ToLower(input), strings.ToLower(original)) {
		lowerInput := strings.ToLower(input)
		lowerOriginal := strings.ToLower(original)

		index := strings.Index(lowerInput, lowerOriginal)
		if index == -1 {
			return input
		}

		// Preserve the case of the replacement
		var result strings.Builder
		result.WriteString(input[:index])

		// Handle capitalization based on the position and context
		if index == 0 {
			// At the beginning - check if original was capitalized
			if len(input) > 0 && isUpperCase(rune(input[0])) {
				result.WriteString(capitalize(replacement))
			} else {
				result.WriteString(replacement)
			}
		} else {
			// In the middle - typically camelCase, so capitalize first letter
			result.WriteString(capitalize(replacement))
		}

		result.WriteString(input[index+len(original):])

		return result.String()
	}

	return input
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}

	return strings.ToUpper(s[:1]) + s[1:]
}

func isUpperCase(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

func isBuiltinIdentifier(name string) bool {
	// Skip Go built-in types and common identifiers
	builtins := map[string]bool{
		// Built-in types
		"bool":       true,
		"byte":       true,
		"complex64":  true,
		"complex128": true,
		"error":      true,
		"float32":    true,
		"float64":    true,
		"int":        true,
		"int8":       true,
		"int16":      true,
		"int32":      true,
		"int64":      true,
		"rune":       true,
		"string":     true,
		"uint":       true,
		"uint8":      true,
		"uint16":     true,
		"uint32":     true,
		"uint64":     true,
		"uintptr":    true,

		// Common interface methods
		"String":  true,
		"Error":   true,
		"Write":   true,
		"Read":    true,
		"Close":   true,
		"Process": true,

		// Go keywords - never replace these
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

		// Common short names already
		"ctx":    true,
		"err":    true,
		"req":    true,
		"res":    true,
		"resp":   true,
		"param":  true,
		"temp":   true,
		"src":    true,
		"dst":    true,
		"db":     true,
		"pwd":    true,
		"usr":    true,
		"srv":    true,
		"svc":    true,
		"len":    true,
		"max":    true,
		"min":    true,
		"buf":    true,
		"img":    true,
		"num":    true,
		"txt":    true,
		"dict":   true,
		"seq":    true,
		"char":   true,
		"ts":     true,
		"pos":    true,
		"ptr":    true,
		"idx":    true,
		"val":    true,
		"ref":    true,
		"orig":   true,
		"addr":   true,
		"pre":    true,
		"cur":    true,
		"init":   true,
		"cnt":    true,
		"pkg":    true,
		"cmd":    true,
		"msg":    true,
		"info":   true,
		"ver":    true,
		"util":   true,
		"calc":   true,
		"obj":    true,
		"arg":    true,
		"v":      true,
		"config": true,
	}

	return builtins[name]
}

func shouldExcludeFile(filename string, config Config) bool {
	for _, pattern := range config.ExcludeFiles {
		matched, err := filepath.Match(pattern, filepath.Base(filename))
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
