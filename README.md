# gonamefix

[![Go Reference](https://pkg.go.dev/badge/github.com/xbpk3t/gonamefix.svg)](https://pkg.go.dev/github.com/xbpk3t/gonamefix)
[![Go Report Card](https://goreportcard.com/badge/github.com/xbpk3t/gonamefix)](https://goreportcard.com/report/github.com/xbpk3t/gonamefix)

`gonamefix` is a Go linter that checks for prohibited naming conventions and suggests replacements. It helps enforce consistent naming patterns across your Go codebase by identifying long variable names that can be shortened according to predefined rules.

## Features

- **Go/analysis based**: Built using the official Go analysis framework
- **Auto-fix support**: Automatically fix naming issues with `-fix` flag
- **Smart camelCase handling**: Properly handles compound words (e.g., `userRequest` → `usrReq`)
- **Keyword protection**: Only blocks exact Go keywords, allows compound words like `forNested`
- **Built-in mappings**: Includes common naming patterns out of the box
- **File/directory exclusion**: Exclude specific files and directories from checks

## Installation

```bash
go install github.com/xbpk3t/gonamefix/cmd/gonamefix@latest
```

## Usage

### Basic Usage

```bash
# Check files for naming issues
gonamefix ./...

# Auto-fix naming issues
gonamefix -fix ./...

# Check specific package
gonamefix ./pkg/mypackage

# Check single file
gonamefix myfile.go
```

## Default Mappings

The linter includes built-in mappings for common long names:

```
request → req          response → res         parameter → param
temporary → temp       source → src           database → db
password → pwd         user → usr             server → srv
service → svc          configuration → config object → obj
argument → arg         variable → v           calculate → calc
maximum → max          minimum → min          address → addr
reference → ref        original → orig        previous → prev
current → cur          buffer → buf           length → len
image → img            number → num           text → txt
dictionary → dict      sequence → seq         character → char
timestamp → ts         position → pos         pointer → ptr
index → idx            value → val            initialize → init
destination → dst      count → cnt            package → pkg
command → cmd          message → msg          information → info
context → ctx          version → ver          utility → util
```

## Examples

### Before

```go
package main

var (
    request    string
    response   []byte
    parameter  int
    temporary  bool
    database   string
    password   string
    user       string
    server     string
)

func processRequest() {}
func handleResponse() {}
func validateParameter() {}

type RequestHandler struct{}
type ResponseWriter struct{}

var userRequest string
var serverResponse []byte

// These are OK - compound words with keywords
var forNested int      // OK
var ifCondition bool   // OK  
var packageInfo string // becomes pkgInfo
```

### After (with -fix)

```go
package main

var (
    req    string
    res    []byte
    param  int
    temp   bool
    db     string
    pwd    string
    usr    string
    srv    string
)

func processReq() {}
func handleRes() {}
func validateParam() {}

type ReqHandler struct{}
type ResWriter struct{}

var userReq string
var serverRes []byte

// These are OK - compound words with keywords
var forNested int      // OK
var ifCondition bool   // OK
var pkgInfo string     // fixed from packageInfo
```

## What Gets Checked

The linter checks the following Go constructs:

- Variable declarations
- Function names and parameters  
- Method names and receivers
- Type names (structs, interfaces, etc.)
- Struct field names
- Function return parameter names

## What Gets Skipped

- Go built-in types (`string`, `int`, `error`, etc.)
- **Exact Go keywords only** (`var`, `func`, `if`, etc.) - compound words like `forNested` are allowed
- Common interface methods (`String`, `Error`, `Write`, etc.)
- Already shortened names (`req`, `res`, `ctx`, etc.)

## Key Improvements

1. **Smart Keyword Handling**: Unlike other linters, `gonamefix` only blocks exact keyword matches. `forNested` is allowed, but `for` alone would be blocked (though it would be a syntax error anyway).

2. **Proper camelCase**: Handles compound words correctly:
   - `userRequest` → `usrReq` (not `usrRequest`)
   - `RequestHandler` → `ReqHandler` (not `reqHandler`)
   - `processUserRequest` → `processUsrReq`

3. **No Configuration Required**: Works out of the box with sensible defaults.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built using the [Go analysis framework](https://golang.org/x/tools/go/analysis)
- Inspired by [forcetypeassert](https://github.com/gostaticanalysis/forcetypeassert) project structure
- Follows Go naming conventions and best practices