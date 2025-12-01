package main

// inputType represents the root structure of the configuration file.
// It contains the suite name, global variables, and the list of tests to execute.
type inputType struct {
	Name      string          `json:"name"`
	Variables variablesStruct `json:"variables"`
	Tests     []test          `json:"tests"`
}

// test defines the configuration for a single integration test step.
// It includes request details (Method, URL, Body), expected outcomes,
// and instructions on data extraction (ToStore).
type test struct {
	Number           int               `json:"num"`
	Method           string            `json:"method"`
	Url              string            `json:"url"`
	Header           map[string]string `json:"header"`
	Body             any               `json:"body"`
	ActualStatus     string            `json:"actual_status"`
	ExpectedStatus   string            `json:"expected_status"`
	ActualResponse   string            `json:"actual_response"`
	ExpectedResponse any               `json:"expected_response"`
	ToStore          map[string]string `json:"var_to_store"`
	TimeTaken        string            `json:"time"`
	Logs             []string          `json:"logs"`
}

// variablesStruct is a map used to store dynamic values during test execution.
// It holds both global configuration variables and values extracted from responses.
type variablesStruct map[string]any

// variables holds the state of all stored values throughout the lifecycle of the application.
var variables = make(variablesStruct)

var t *test

// flags
// output_dir points to directory for reports
var output_dir *string

// path represents path of the test config json file
var path *string

// template refers to template.html file path from which reports are generated
var template_file *string
