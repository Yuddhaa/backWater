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

func main() {
	// Parse command line flags
	path = flag.String("path", "./test.json", "path of the test json file")
	output_dir = flag.String("output_dir", "./reports", "directory path for the report. Default: ./reports")
	template_file = flag.String("template", "./template.html", "template refers to template.html file path from which reports are generated. Default: ./template.html")
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

	total = len(input.Tests)
	fmt.Printf("\n\t--- Total Number of Tests:%v ---\n\n", total)

	start := time.Now()

	// 3. Iterate through and execute tests
	for i := range input.Tests {
		// Use a pointer to the current test so updates (Url, Logs, etc.) are reflected directly
		t = &input.Tests[i]
		// Test = &input.Tests[i]

		testStart := time.Now()
		testNo += 1

		// Set the test number in the struct if not present (optional, but good for reporting)
		t.Number = testNo

		LogMsg("\n------------- Test %d: [%s] %s -------------\n\n", testNo, t.Method, t.Url)

		// --- Variable Substitution & Pre-processing ---
		if ok := t.preProcess(testNo); !ok {
			failed++
			continue
		}

		var body io.Reader

		// --- Request Construction ---

		// 3.1 Convert body into io.Reader if body exists
		if t.Body != nil {
			jsonData, err := json.Marshal(t.Body)
			if err != nil {
				failed++
				LogMsg("[FAIL] %v: Invalid JSON body in test config: %v\n\n", testNo, err)
				LogMsg("------------- Test %v Completed-------------\n\n", testNo)
				continue
			}
			body = bytes.NewBuffer(jsonData)
		}

		// 3.2 Create new request
		req, err = http.NewRequest(t.Method, t.Url, body)
		if err != nil {
			failed++
			LogMsg("[FAIL] %v: Could not create request: %v\n\n", testNo, err)
			LogMsg("------------- Test %v Completed-------------\n\n", testNo)
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		// 3.3 Set custom headers if present
		if t.Header != nil {
			for k, v := range t.Header {
				req.Header.Set(k, v)
			}
		}

		// --- Execution ---

		// Do the http call
		res, err := client.Do(req)
		if err != nil {
			failed++
			LogMsg("[FAIL] %v: Network error: %v\n\n", testNo, err)
			LogMsg("------------- Test %v Completed-------------\n\n", testNo)
			continue
		}

		// Read the actual response body
		actualBody, err := io.ReadAll(res.Body)
		if err != nil {
			failed++
			LogMsg("[FAIL] %v: Failed to process actual body.\n\n", testNo)
			LogMsg("------------- Test %v Completed-------------\n\n", testNo)
			res.Body.Close()
			continue
		}
		t.ActualStatus = res.Status
		t.ActualResponse = string(actualBody)

		// --- Validation ---
		// 1. Status Check
		statusMatch := false
		if res.Status != t.ExpectedStatus {
			LogMsg("[FAIL] %v: Status Mismatch.\n\tExpected: %s\n\tGot:      %s\n", testNo, t.ExpectedStatus, res.Status)
		} else {
			statusMatch = true
			LogMsg("[PASS] Status OK.\n")
		}

		bodyMatch := true
		if !statusMatch {
			// 2. Body Check (Subset Validation)
			if t.ExpectedResponse != nil {
				// Unmarshal actual body to 'any' for comparison
				var actualJSON any
				if err := json.Unmarshal(actualBody, &actualJSON); err != nil {
					bodyMatch = false
					LogMsg("[FAIL] Response body is not valid JSON, cannot validate against expected.\n")
				} else {
					if validateBody(t.ExpectedResponse, actualJSON) {
						LogMsg("[PASS] Body Subset Match OK.\n")
					} else {
						bodyMatch = false
						LogMsg("[FAIL] Body Mismatch.\n")
					}
				}
			}
		} else {
			LogMsg("[NOTE] since status did not match, skipping body validation.")
		}

		if !statusMatch || !bodyMatch {
			failed++
		} else {
			passed++
		}
		// // Status validation
		// if res.Status != t.ExpectedStatus {
		// 	failed++
		// 	LogMsg("[FAIL] %v: Status Mismatch.\n\tExpected: %s\n\tGot:      %s\n", testNo, t.ExpectedStatus, res.Status)
		// } else {
		// 	passed++
		// 	LogMsg("[PASS] Status OK.\n")
		// }
		//
		// Store required body variables
		if ok := storeBodyVariables(testNo, actualBody, variables, t.ToStore); !ok {
			LogMsg("[NOTE] %v: Failed to store body variables.\n\n", testNo)
		} else if len(t.ToStore) > 0 {
			LogMsg("Variables stored successfully.\n")
		}

		// Cleanup
		// io.Copy(io.Discard, res.Body)
		res.Body.Close()
		t.TimeTaken = time.Since(testStart).String()
		LogMsg("Test took %v\n", t.TimeTaken)
		LogMsg("------------- Test %v Completed-------------\n\n", testNo)
	}

	totalAllTestTime := time.Since(start).String()

	// Final Report
	fmt.Println("------------------- Test Ended -------------------")
	fmt.Printf("\nTotal Number of Tests:%v\n", total)
	fmt.Printf("Passed: %v\n", passed)
	fmt.Printf("Failed: %v\n", failed)
	fmt.Printf("Total time elapsed:%v\n", totalAllTestTime)

	GenerateHTMLReport(input, totalAllTestTime, total, passed)
}

// LogMsg prints to console and appends to the logs of the specific test case.
// It accepts a pointer to the test struct so it can be called from other files (store.go, process.go).
func LogMsg(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Print(msg) // Print to standard output
	if t != nil {
		t.Logs = append(t.Logs, msg)
	}
}

// Helper to print indented JSON (not used in main loop anymore, but kept for util)
func printIndentJson(s string, v any) {
	temp, _ := json.MarshalIndent(v, "", "    ")
	fmt.Printf("%v: %v\n", s, string(temp))
}
