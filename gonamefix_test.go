package gonamefix

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()

	// Test with custom configuration - this should match what's expected in testdata/src/a/a.go
	config := Config{
		Check: [][]string{
			{"request", "req"},
			{"response", "res"},
			{"parameter", "param"},
			{"temporary", "temp"},
			{"source", "src"},
			{"database", "db"},
			{"password", "pwd"},
			{"user", "usr"},
			{"server", "srv"},
			{"service", "svc"},
			{"configuration", "config"},
			{"package", "pkg"},
		},
		ExcludeFiles:  []string{"*.pb.go", "*_test.go"},
		ExcludeDirs:   []string{"vendor", "node_modules", ".git"},
		CaseSensitive: false,
	}

	analyzer := NewAnalyzer(config)
	analysistest.Run(t, testdata, analyzer, "a")
}

func TestAnalyzerNoMappings(t *testing.T) {
	testdata := analysistest.TestData()

	// Test with no mappings - should not report any issues
	config := Config{
		Check:         [][]string{},
		ExcludeFiles:  []string{"*.pb.go", "*_test.go"},
		ExcludeDirs:   []string{"vendor", "node_modules", ".git"},
		CaseSensitive: false,
	}

	analyzer := NewAnalyzer(config)
	// This should pass without any reports since no mappings are configured
	analysistest.Run(t, testdata, analyzer, "b")
}

func TestAnalyzerCaseSensitive(t *testing.T) {
	testdata := analysistest.TestData()

	// Test with case sensitive configuration
	config := Config{
		Check: [][]string{
			{"request", "req"},
			{"Request", "Req"}, // Case sensitive mapping
		},
		ExcludeFiles:  []string{"*.pb.go", "*_test.go"},
		ExcludeDirs:   []string{"vendor", "node_modules", ".git"},
		CaseSensitive: true,
	}

	analyzer := NewAnalyzer(config)
	analysistest.Run(t, testdata, analyzer, "c") // Use c.go which has expected diagnostics for case sensitive
}

func TestConfigFunctions(t *testing.T) {
	// Test buildNameMappings
	mappings := buildNameMappings([][]string{
		{"request", "req"},
		{"response", "res"},
		{"invalid"}, // Should be ignored
	})

	if len(mappings) != 2 {
		t.Errorf("Expected 2 mappings, got %d", len(mappings))
	}

	if mappings["request"] != "req" {
		t.Errorf("Expected request->req mapping")
	}

	if mappings["response"] != "res" {
		t.Errorf("Expected response->res mapping")
	}

	// Test buildPatterns
	patterns := buildPatterns(mappings, false)
	if len(patterns) != 2 {
		t.Errorf("Expected 2 patterns, got %d", len(patterns))
	}

	// Test case sensitive patterns
	patterns = buildPatterns(mappings, true)
	if len(patterns) != 2 {
		t.Errorf("Expected 2 patterns for case sensitive, got %d", len(patterns))
	}
}

func TestReplaceInName(t *testing.T) {
	tests := []struct {
		testName      string
		inputName     string
		original      string
		replacement   string
		caseSensitive bool
		expected      string
	}{
		{"exact match", "request", "request", "req", false, "req"},
		{"case insensitive exact", "Request", "request", "req", false, "Req"},
		{"camelCase", "processRequest", "request", "req", false, "processReq"},
		{"PascalCase", "ProcessRequest", "request", "req", false, "ProcessReq"},
		{"no match", "something", "request", "req", false, "something"},
		{"case sensitive exact", "request", "request", "req", true, "req"},
		{"case sensitive no match", "Request", "request", "req", true, "Request"},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			result := replaceInName(tt.inputName, tt.original, tt.replacement, tt.caseSensitive)
			if result != tt.expected {
				t.Errorf("replaceInName(%q, %q, %q, %t) = %q, want %q",
					tt.inputName, tt.original, tt.replacement, tt.caseSensitive, result, tt.expected)
			}
		})
	}
}

func TestIsGoKeyword(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"func", true},
		{"var", true},
		{"if", true},
		{"request", false},
		{"notakeyword", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isGoKeyword(tt.name)
			if result != tt.expected {
				t.Errorf("isGoKeyword(%q) = %t, want %t", tt.name, result, tt.expected)
			}
		})
	}
}

func TestShouldExcludeFile(t *testing.T) {
	config := Config{
		ExcludeFiles: []string{"*.pb.go", "*_test.go"},
		ExcludeDirs:  []string{"vendor", "node_modules"},
	}

	tests := []struct {
		filename string
		expected bool
	}{
		{"test.pb.go", true},
		{"test_test.go", true},
		{"vendor/some/file.go", true},
		{"node_modules/package/file.go", true},
		{"normal.go", false},
		{"main.go", false},
		{"/path/to/test.pb.go", true},
		{"/path/to/normal.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := shouldExcludeFile(tt.filename, config)
			if result != tt.expected {
				t.Errorf("shouldExcludeFile(%q) = %t, want %t", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestEdgeCases(t *testing.T) {
	// Test with empty strings and nil values
	result := replaceInName("", "request", "req", false)
	if result != "" {
		t.Errorf("replaceInName with empty name should return empty string")
	}

	result = replaceInName("test", "", "req", false)
	if result != "test" {
		t.Errorf("replaceInName with empty original should return original name")
	}

	// Test buildNameMappings with invalid data
	mappings := buildNameMappings([][]string{
		{"valid", "mapping"},
		{},                          // empty slice
		{"single"},                  // only one element
		{"too", "many", "elements"}, // too many elements
	})

	if len(mappings) != 1 {
		t.Errorf("Expected 1 valid mapping, got %d", len(mappings))
	}

	// Test camelCase edge cases
	tests := []struct {
		input       string
		original    string
		replacement string
		expected    string
	}{
		{"requestData", "request", "req", "reqData"},
		{"DataRequest", "request", "req", "DataReq"},
		{"MyRequestHandler", "request", "req", "MyReqHandler"},
		{"requestRequestRequest", "request", "req", "reqRequestRequest"}, // Should only replace first
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := replaceInName(tt.input, tt.original, tt.replacement, false)
			if result != tt.expected {
				t.Errorf("replaceInName(%q, %q, %q, false) = %q, want %q",
					tt.input, tt.original, tt.replacement, result, tt.expected)
			}
		})
	}
}
