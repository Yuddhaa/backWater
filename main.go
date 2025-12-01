package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// main is the entry point of the test runner.
// It orchestrates the loading of configuration, processing of variables,
// execution of HTTP requests, and validation of results.
func main() {
	// Parse command line flags
	path := flag.String("path", "./test.json", "path of the test json file. Default value: ./test.json")
	flag.Parse()

	// Initialize execution variables
	var input inputType
	client := http.Client{}
	testNo := 0
	var req *http.Request
	var total, passed, failed int

	fmt.Println("------------------- Test Started -------------------")

	// 1. Open the configuration file
	file, err := os.Open(*path)
	if err != nil {
		log.Fatalf("cannot read the contents of test.json.\nErr:%v", err)
	}
	defer file.Close()

	// 2. Decode the JSON content into the struct
	if err := json.NewDecoder(file).Decode(&input); err != nil {
		log.Fatalf("cannot decode test.json.\nErr:%v", err)
	}
	fmt.Printf("\n\t--- Name: %v ---\n", input.Name)

	// Load global variables defined in the config
	storeGlobalVariables(variables, input.Variables)
	// printIndentJson("Global variables", variables)

	total = len(input.Tests)
	fmt.Printf("\n\t--- Total Number of Tests:%v ---\n\n", total)

	start := time.Now()
	// 3. Iterate through and execute tests
	for i, test := range input.Tests {
		testStart := time.Now()
		testNo += 1
		fmt.Printf("\n------------- Test %d: [%s] %s -------------\n\n", testNo, test.Method, test.Url)

		// --- Variable Substitution & Pre-processing ---

		// Process Headers
		// printIndentJson("header before processing", test.Header)
		if ok := processHeader(test.Header); !ok {
			failed++
			fmt.Printf("[FAIL] %v. Failed to process header.\n\n", testNo)
			fmt.Printf("------------- Test %v Completed-------------\n\n", testNo)
			continue
		}
		// printIndentJson("header after processing", test.Header)

		// Process URL (substitute variables in the path)
		// printIndentJson("url after processing", test.Url)
		var ok bool
		if test.Url, ok = processUrl(test.Url); !ok {
			failed++
			fmt.Printf("[FAIL] %v. Failed to process Url.\n\n", testNo)
			fmt.Printf("------------- Test %v Completed-------------\n\n", testNo)
			continue
		}
		// printIndentJson("url before processing", test.Url)

		// Process Expected Response (substitute variables for validation)
		// printIndentJson("expected_response before processing", test.ExpectedResponse)
		if ok := processBody(test.ExpectedResponse); !ok {
			failed++
			fmt.Printf("[FAIL] %v. Failed to process expected_response.\n\n", testNo)
			fmt.Printf("------------- Test %v Completed-------------\n\n", testNo)
			continue
		}
		// printIndentJson("expected_response after processing", test.ExpectedResponse)

		// Process Request Body (substitute variables in the payload)
		// printIndentJson("body before processing", test.Body)
		if ok := processBody(test.Body); !ok {
			failed++
			fmt.Printf("[FAIL] %v. Failed to process Body.\n\n", testNo)
			fmt.Printf("------------- Test %v Completed-------------\n\n", testNo)
			continue
		}
		// printIndentJson("body after processing", test.Body)

		var body io.Reader

		// --- Request Construction ---

		// 3.1 Convert body into io.Reader if body exists
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

		// 3.2 Create new request
		req, err = http.NewRequest(test.Method, test.Url, body)
		if err != nil {
			failed++
			fmt.Printf("[FAIL] %v Could not create request: %v\n\n", testNo, err)
			fmt.Printf("------------- Test %v Completed-------------\n\n", testNo)
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		// 3.3 Set custom headers if present
		if test.Header != nil {
			for k, v := range test.Header {
				req.Header.Set(k, v)
			}
		}

		// --- Execution ---

		// Do the http call
		res, err := client.Do(req)
		if err != nil {
			failed++
			fmt.Printf("[FAIL] %v Network error: %v\n\n", testNo, err)
			fmt.Printf("------------- Test %v Completed-------------\n\n", testNo)
			continue
		}

		// Read the actual response body
		actualBody, err := io.ReadAll(res.Body)
		if err != nil {
			failed++
			fmt.Printf("[FAIL] %v. Failed to process actual body.\n\n", testNo)
			fmt.Printf("------------- Test %v Completed-------------\n\n", testNo)
			res.Body.Close()
			continue
		}
		input.Tests[i].ActualStatus = res.Status
		input.Tests[i].ActualResponse = string(actualBody)

		// --- Validation ---

		// Status validation
		if res.Status != test.ExpectedStatus {
			// fmt.Printf("------------- Test %v Failed-------------", testNo)
			failed++
			fmt.Printf("[FAIL] %v Status Mismatch.\n\tExpected: %s\n\tGot:      %s\n", testNo, test.ExpectedStatus, res.Status)
		} else {
			passed++
			fmt.Printf("[PASS] Status OK.\n")
		}

		// Store required body variables for future tests
		if ok := storeBodyVariables(testNo, actualBody, variables, test.ToStore); !ok {
			fmt.Printf("[NOTE] %v. Failed to store body variables.\n\n", testNo)
		}

		// Cleanup
		io.Copy(io.Discard, res.Body)
		res.Body.Close()
		testTime := time.Since(testStart).String()
		input.Tests[i].TimeTaken = testTime
		fmt.Printf("Test took %v\n", testTime)
		fmt.Printf("------------- Test %v Completed-------------\n\n", testNo)
	}
	totalAllTestTime := time.Since(start).String()

	// Final Report
	fmt.Println("------------------- Test Ended -------------------")
	fmt.Printf("\nTotal Number of Tests:%v\n", total)
	fmt.Printf("Passed: %v\n", passed)
	fmt.Printf("Failed: %v\n", failed)
	fmt.Printf("Total time elapsed:%v\n", totalAllTestTime)

	GenerateHTMLReport(input, totalAllTestTime, total, passed, "report.html")
	// printIndentJson("all variables", variables)
}

// printIndentJson formats and prints a data structure as indented JSON for debugging purposes.
func printIndentJson(s string, v any) {
	temp, _ := json.MarshalIndent(v, "", "    ")
	fmt.Printf("%v: %v\n", s, string(temp))
}
