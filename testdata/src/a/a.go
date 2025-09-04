package a

// Test basic variable names
var (
	request   string // want "suggest replacing 'request' with 'req'"
	response  []byte // want "suggest replacing 'response' with 'res'"
	parameter int    // want "suggest replacing 'parameter' with 'param'"
	temporary bool   // want "suggest replacing 'temporary' with 'temp'"
	source    string // want "suggest replacing 'source' with 'src'"
	database  string // want "suggest replacing 'database' with 'db'"
	password  string // want "suggest replacing 'password' with 'pwd'"
	user      string // want "suggest replacing 'user' with 'usr'"
	server    string // want "suggest replacing 'server' with 'srv'"
	service   string // want "suggest replacing 'service' with 'svc'"
)

// Test function names
func processRequest() {} // want "suggest replacing 'processRequest' with 'processReq'"

func handleResponse() {} // want "suggest replacing 'handleResponse' with 'handleRes'"

func validateParameter() {} // want "suggest replacing 'validateParameter' with 'validateParam'"

// Test function parameters
func processData(request string, response []byte) { // want "suggest replacing 'request' with 'req'" "suggest replacing 'response' with 'res'"
	// Function body
}

// Test type names
type RequestHandler struct{} // want "suggest replacing 'RequestHandler' with 'ReqHandler'"

type ResponseWriter struct{} // want "suggest replacing 'ResponseWriter' with 'ResWriter'"

// Test struct fields
type Config struct {
	request  string // want "suggest replacing 'request' with 'req'"
	response string // want "suggest replacing 'response' with 'res'"
}

// Test cases that should NOT be flagged (already short or keywords)
var req string // OK - already short
var res []byte // OK - already short
var pkg string // OK - already short
var ctx string // OK - already short

// Test compound words with keywords - these should be allowed
var forNested int      // OK - compound with keyword
var ifCondition bool   // OK - compound with keyword
var packageInfo string // want "suggest replacing 'packageInfo' with 'pkgInfo'"

func testBasic() {
	_ = 1 // avoid unused warnings
}
