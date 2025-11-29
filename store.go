package main

import (
	"encoding/json"
	"fmt"
	"maps"
	"strconv"
	"strings"
)

// storeGlobalVariables merges the variables defined in the input configuration
// into the main global variable store. It uses maps.Copy to perform a shallow merge.
func storeGlobalVariables(variables, input variablesStruct) {
	maps.Copy(variables, input)
}

// storeBodyVariables extracts specific values from the HTTP response body based on
// the 'toStore' map configuration. It namespaces the extracted variables with the
// test number (e.g., "test_1_varName") and saves them to the global variables map.
// Returns false if extraction fails or JSON is invalid.
func storeBodyVariables(testNo int, body []byte, variables variablesStruct, toStore map[string]string) bool {
	// 1. Basic validation to ensure the body looks like a JSON object or array
	if len(body) == 0 || body[0] != '{' && body[0] != '[' {
		fmt.Println("Body is not JSON, skipping unmarshal.")
		return true
	}
	// fmt.Println("body:", string(body))

	// 2. Unmarshal body into a generic map for traversal
	var tempBody map[string]any
	if err := json.Unmarshal(body, &tempBody); err != nil {
		fmt.Println("Error in Unmarshal of the body. Err:" + err.Error())
		return false
	}

	// 3. Iterate over the requested variables to extract
	for k, v := range toStore {
		// Construct the namespaced key: "test_{testNo}_{variableName}"
		n := strconv.FormatInt(int64(testNo), 10)
		varValue, ok := getNestedValue(v, tempBody)
		if !ok {
			return false
		}
		variables["test_"+n+"_"+k] = varValue
	}
	return true
}

// getNestedValue retrieves a value from a generic JSON structure using a dot-notation path.
// It supports array indexing (e.g., "data.items[0].id").
// Returns the found value and a boolean indicating success.
func getNestedValue(path string, data any) (any, bool) {
	keys := strings.Split(path, ".")
	current := data

	// Traverse the data structure segment by segment
	for _, key := range keys {
		// Check for array notation (e.g., "items[0]")
		iIdx := strings.Index(key, "[") + 1
		var i int
		var err error
		isArr := false

		// Parse array index if present
		if iIdx > 0 {
			i, err = strconv.Atoi(string(key[iIdx]))
			if err != nil {
				fmt.Printf("Invalid index in path: %v. Error:%v\n", path, err.Error())
				return nil, false
			}
			// Strip the array notation from the key name for map lookup
			key = strings.Split(key, "[")[0]
			isArr = true
		}

		// Ensure current node is a map
		m, mok := current.(map[string]any)
		if !mok {
			fmt.Printf("path segment '%s' not found or parent is not a map\n", key)
			return nil, false
		}

		// Look up the key in the map
		val, exists := m[key]
		if !exists {
			fmt.Printf("key '%s' not found\n", key)
			return nil, false
		}

		// Handle array navigation if the key implied an array
		if isArr {
			fmt.Println("inside is arr if")
			arr, ok := val.([]any)
			if !ok {
				fmt.Printf("path segment '%s' is not an array\n", key)
				return nil, false
			}
			// Update current pointer to the specific array element
			current = arr[i]
		} else {
			// Update current pointer to the map value
			current = val
		}
	}
	return current, true
}
