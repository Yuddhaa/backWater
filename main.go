package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type inputType struct {
	Name      string          `json:"name"`
	Variables variablesStruct `json:"variables"`
	Tests     []test          `json:"tests"`
}

type test struct {
	Number           int               `json:"num"`
	Method           string            `json:"method"`
	Url              string            `json:"url"`
	Header           map[string]string `json:"header"`
	Body             any               `json:"body"`
	ExpectedStatus   string            `json:"expected_status"`
	ExpectedResponse any               `json:"expected_response"`
	ToStore          map[string]string `json:"var_to_store"`
}

type variablesStruct map[string]any

var variables = make(variablesStruct)

func main() {
	// 0. Variables
	var input inputType
	client := http.Client{}
	testNo := 0
	var req *http.Request
	var total, passed, failed int

	fmt.Println("------------------- Test Started -------------------")
	// 1. open the json file
	file, err := os.Open("./test.json")
	if err != nil {
		log.Fatalf("cannot read the contents of test.json.\nErr:%v", err)
	}
	defer file.Close()
	// 2. open the json file
	if err := json.NewDecoder(file).Decode(&input); err != nil {
		log.Fatalf("cannot decode test.json.\nErr:%v", err)
	}
	fmt.Printf("\n\t--- Name: %v ---\n", input.Name)

	// Presetup of the tests
	storeGlobalVariables(variables, input.Variables)
	// printIndentJson("Global variables", variables)

	total = len(input.Tests)
	fmt.Printf("\n\t--- Total Number of Tests:%v ---\n\n", total)
	// 3. do the testing
	for _, test := range input.Tests {
		testNo += 1
		fmt.Printf("\n------------- Test %d: [%s] %s -------------\n\n", testNo, test.Method, test.Url)

		// Pre processing of the tests
		// printIndentJson("header before processing", test.Header)
		if ok := processHeader(test.Header); !ok {
			failed++
			fmt.Printf("[FAIL] %v. Failed to process header.\n\n", testNo)
			fmt.Printf("------------- Test %v Completed-------------\n\n", testNo)
			continue
		}
		// printIndentJson("header after processing", test.Header)

		// process url
		// printIndentJson("url after processing", test.Url)
		var ok bool
		if test.Url, ok = processUrl(test.Url); !ok {
			failed++
			fmt.Printf("[FAIL] %v. Failed to process header.\n\n", testNo)
			fmt.Printf("------------- Test %v Completed-------------\n\n", testNo)
			continue
		}
		// printIndentJson("url before processing", test.Url)

		var body io.Reader
		// 3.1 convert body into io.Reader
		if test.Body != nil {
			// 3.1.1. Convert the generic 'any' back into JSON bytes
			jsonData, err := json.Marshal(test.Body)
			if err != nil {
				failed++
				fmt.Printf("[FAIL] %v. Invalid JSON body in test config: %v\n\n", testNo, err)
				fmt.Printf("------------- Test %v Completed-------------\n\n", testNo)
				continue
			}
			// 3.1.2 Create a Reader from those bytes
			body = bytes.NewBuffer(jsonData)
		}

		// 3.2 new request and set the header
		req, err = http.NewRequest(test.Method, test.Url, body)
		if err != nil {
			failed++
			fmt.Printf("[FAIL] %v Could not create request: %v\n\n", testNo, err)
			fmt.Printf("------------- Test %v Completed-------------\n\n", testNo)
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		// 3.3 set header if present
		if test.Header != nil {
			for k, v := range test.Header {
				req.Header.Set(k, v)
			}
		}

		// Do the http call
		res, err := client.Do(req)
		if err != nil {
			failed++
			fmt.Printf("[FAIL] %v Network error: %v\n\n", testNo, err)
			fmt.Printf("------------- Test %v Completed-------------\n\n", testNo)
			continue
		}

		// Read the body
		actualBody, err := io.ReadAll(res.Body)
		if err != nil {
			failed++
			fmt.Printf("[FAIL] %v. Failed to process actual body.\n\n", testNo)
			fmt.Printf("------------- Test %v Completed-------------\n\n", testNo)
			res.Body.Close()
			continue
		}

		// Status validation
		if res.Status != test.ExpectedStatus {
			// fmt.Printf("------------- Test %v Failed-------------", testNo)
			failed++
			fmt.Printf("[FAIL] %v Status Mismatch.\n\tExpected: %s\n\tGot:      %s\n", testNo, test.ExpectedStatus, res.Status)
		} else {
			passed++
			fmt.Printf("[PASS] Status OK.\n")
		}

		// Store required body variables
		if ok := storeBodyVariables(testNo, actualBody, variables, test.ToStore); !ok {
			fmt.Printf("[NOTE] %v. Failed to store body variables.\n\n", testNo)
		}

		io.Copy(io.Discard, res.Body)
		res.Body.Close()
		fmt.Printf("------------- Test %v Completed-------------\n\n", testNo)
	}

	fmt.Println("------------------- Test Ended -------------------")
	fmt.Printf("\nTotal Number of Tests:%v\n", total)
	fmt.Printf("Passed: %v\n", passed)
	fmt.Printf("Failed: %v\n", failed)

	printIndentJson("all variables", variables)
}

func printIndentJson(s string, v any) {
	temp, _ := json.MarshalIndent(v, "", "    ")
	fmt.Printf("%v: %v\n", s, string(temp))
}
