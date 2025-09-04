package c

// Test file for case sensitive mappings
var (
	request string // want "suggest replacing 'request' with 'req'"
)

// Test function names
func processRequest() {} // want "suggest replacing 'processRequest' with 'processReq'"

// Test function parameters
func processData(request string) { // want "suggest replacing 'request' with 'req'"
	// Function body
}

// Test type names
type RequestHandler struct{} // want "suggest replacing 'RequestHandler' with 'ReqHandler'"

// Test struct fields
type Config struct {
	request string // want "suggest replacing 'request' with 'req'"
}

func testBasic() {
	_ = 1 // avoid unused warnings
}
