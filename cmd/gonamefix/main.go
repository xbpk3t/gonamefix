package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/xbpk3t/gonamefix"
)

var (
	checkFlag         = flag.String("check", "", "Name mappings in format 'old1:new1,old2:new2'")
	excludeFilesFlag  = flag.String("exclude-files", "*.pb.go,*_test.go", "File patterns to exclude")
	excludeDirsFlag   = flag.String("exclude-dirs", "vendor,node_modules,.git", "Directory patterns to exclude")
	caseSensitiveFlag = flag.Bool("case-sensitive", false, "Case sensitive matching")
	recursiveFlag     = flag.Bool("recursive", false, "Recursively scan directories")
	configFileFlag    = flag.String("config", "", "Configuration file path")
	helpFlag          = flag.Bool("help", false, "Show help")
)

func main() {
	flag.Parse()

	if *helpFlag {
		showHelp()
		return
	}

	config, err := loadConfiguration()
	if err != nil {
		log.Fatal(err)
	}

	// If no check mappings provided, show help
	if len(config.Check) == 0 {
		fmt.Println("Error: No name mappings provided.")
		fmt.Println()
		showHelp()
		os.Exit(1)
	}

	analyzer := gonamefix.NewAnalyzer(config)

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("Error: No files or directories specified.")
		showHelp()
		os.Exit(1)
	}

	// Check if we're processing directories or files
	var files []string
	for _, arg := range args {
		if info, err := os.Stat(arg); err == nil && info.IsDir() {
			if *recursiveFlag {
				dirFiles, err := findGoFiles(arg)
				if err != nil {
					log.Printf("Error scanning directory %s: %v", arg, err)
					continue
				}
				files = append(files, dirFiles...)
			} else {
				dirFiles, err := findGoFilesInDir(arg)
				if err != nil {
					log.Printf("Error scanning directory %s: %v", arg, err)
					continue
				}
				files = append(files, dirFiles...)
			}
		} else {
			files = append(files, arg)
		}
	}

	if len(files) == 0 {
		fmt.Println("No Go files found to analyze.")
		return
	}

	// Process each file
	exitCode := 0
	for _, file := range files {
		if err := analyzeFile(analyzer, file); err != nil {
			log.Printf("Error analyzing %s: %v", file, err)
			exitCode = 1
		}
	}

	if exitCode != 0 {
		os.Exit(exitCode)
	}
}

func loadConfiguration() (gonamefix.Config, error) {
	config := gonamefix.Config{
		ExcludeFiles:  strings.Split(*excludeFilesFlag, ","),
		ExcludeDirs:   strings.Split(*excludeDirsFlag, ","),
		CaseSensitive: *caseSensitiveFlag,
	}

	// Parse check flag
	if *checkFlag != "" {
		pairs := strings.Split(*checkFlag, ",")
		for _, pair := range pairs {
			parts := strings.Split(pair, ":")
			if len(parts) == 2 {
				config.Check = append(config.Check, []string{strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])})
			} else {
				return config, fmt.Errorf("invalid mapping format: %s (expected 'old:new')", pair)
			}
		}
	}

	return config, nil
}

func analyzeFile(analyzer *analysis.Analyzer, filename string) error {
	fset := token.NewFileSet()

	// Parse the file
	file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	// Create a pass for the analyzer
	pass := &analysis.Pass{
		Analyzer: analyzer,
		Fset:     fset,
		Files:    []*ast.File{file},
		Report: func(d analysis.Diagnostic) {
			pos := fset.Position(d.Pos)
			fmt.Printf("%s:%d:%d: %s\n", pos.Filename, pos.Line, pos.Column, d.Message)
		},
		ResultOf: make(map[*analysis.Analyzer]interface{}),
	}

	// Need to handle required analyzers
	for _, req := range analyzer.Requires {
		result, err := req.Run(pass)
		if err != nil {
			return fmt.Errorf("required analyzer %s failed: %w", req.Name, err)
		}
		pass.ResultOf[req] = result
	}

	// Run the analyzer
	_, err = analyzer.Run(pass)
	return err
}

func findGoFiles(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".go") && !strings.Contains(path, "vendor/") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func findGoFilesInDir(dir string) ([]string, error) {
	var files []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}
	return files, nil
}

func showHelp() {
	fmt.Println("gonamefix - Go naming convention fixer")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  gonamefix [flags] <files or directories>")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -check string")
	fmt.Println("        Name mappings in format 'old1:new1,old2:new2'")
	fmt.Println("        Example: -check 'request:req,response:res,configuration:config'")
	fmt.Println()
	fmt.Println("  -exclude-files string")
	fmt.Println("        File patterns to exclude (default \"*.pb.go,*_test.go\")")
	fmt.Println()
	fmt.Println("  -exclude-dirs string")
	fmt.Println("        Directory patterns to exclude (default \"vendor,node_modules,.git\")")
	fmt.Println()
	fmt.Println("  -case-sensitive")
	fmt.Println("        Case sensitive matching (default false)")
	fmt.Println()
	fmt.Println("  -recursive")
	fmt.Println("        Recursively scan directories (default false)")
	fmt.Println()
	fmt.Println("  -help")
	fmt.Println("        Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Check single file")
	fmt.Println("  gonamefix -check 'request:req,response:res' file.go")
	fmt.Println()
	fmt.Println("  # Check directory (non-recursive)")
	fmt.Println("  gonamefix -check 'request:req,response:res' ./cmd")
	fmt.Println()
	fmt.Println("  # Check directory recursively")
	fmt.Println("  gonamefix -check 'request:req,response:res' -recursive ./")
	fmt.Println()
	fmt.Println("  # Check multiple files")
	fmt.Println("  gonamefix -check 'request:req,response:res' file1.go file2.go")
}
