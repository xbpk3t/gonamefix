package b

// Test file for no mappings - should not report any issues
var (
	request   string // OK - no mappings configured
	response  []byte // OK - no mappings configured
	parameter int    // OK - no mappings configured
	temporary bool   // OK - no mappings configured
	source    string // OK - no mappings configured
	database  string // OK - no mappings configured
)

// Test function names - should not report any issues
func processRequest() {} // OK - no mappings configured

func handleResponse() {} // OK - no mappings configured

// Test function parameters - should not report any issues
func processData(request string, response []byte) { // OK - no mappings configured
	// Function body
}

// Test type names - should not report any issues
type RequestHandler struct{} // OK - no mappings configured

type ResponseWriter struct{} // OK - no mappings configured

// Test struct fields - should not report any issues
type Config struct {
	request  string // OK - no mappings configured
	response string // OK - no mappings configured
}

func testBasic() {
	_ = 1 // avoid unused warnings
}
