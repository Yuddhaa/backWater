package main

import (
	"bytes"
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
func storeBodyVariables(testNo int, body []byte, variables map[string]any, toStore map[string]string) bool {
	if toStore != nil {
		return true
	}
	// 1. Basic validation
	body = bytes.TrimSpace(body)
	if len(body) == 0 {
		LogMsg("Body is empty, skipping.")
		return true
	}
	// Check if it looks like JSON (starts with { or [)
	if body[0] != '{' && body[0] != '[' {
		LogMsg("While storing variables, body does not look like JSON (starts with '%c'), so skipping storing of variables.\n", body[0])
		return true
	}

	// 2. Unmarshal into 'any'.
	// This handles both Objects (map[string]any) and Arrays ([]any) automatically.
	var bodyData any
	if err := json.Unmarshal(body, &bodyData); err != nil {
		LogMsg("Error in Unmarshal of the body. Err: %v\n", err)
		return false
	}

	// 3. Iterate over the requested variables to extract
	success := true
	for k, v := range toStore {
		// Construct the namespaced key: "test_{testNo}_{variableName}"
		// using Sprintf is cleaner than FormatInt manual concatenation
		keyName := fmt.Sprintf("test_%d_%s", testNo, k)

		// We just pass the generic bodyData.
		// getNestedValue is smart enough to handle maps vs arrays.
		varValue, ok := getNestedValue(v, bodyData)
		if !ok {
			LogMsg("Failed to extract '%s' (path: %s).\n All the tests referencing this variable might fail.\n", k, v)
			// We continue so we can try to find other variables even if one fails
			continue
		}

		variables[keyName] = varValue
		// Optional: Debug log
		LogMsg("[NOTE] Stored %s = %v\n", keyName, varValue)
	}

	return success
}

// getNestedValue retrieves a value from a generic JSON structure using a dot-notation path.
// It supports:
// 1. Standard keys: "user.name"
// 2. Array indices: "users[0].id"
// 3. Multi-digit indices: "data[100]"
// 4. Nested arrays: "grid[0][1]"
// 5. Root arrays: "[0].name"
func getNestedValue(path string, data any) (any, bool) {
	if path == "" {
		return data, true
	}

	// 1. Split the path by dots (standard JSON traversal)
	segments := strings.Split(path, ".")
	current := data

	for _, segment := range segments {
		// Check if this segment contains an array notation like "items[0]"
		bracketIdx := strings.Index(segment, "[")

		// --- CASE A: Simple Map Key (e.g., "name") ---
		if bracketIdx == -1 {
			m, ok := current.(map[string]any)
			if !ok {
				LogMsg("Path '%s' failed at segment '%s': current value is not a map (got type %T)\n", path, segment, current)
				return nil, false // Current node is not a map
			}
			val, exists := m[segment]
			if !exists {
				LogMsg("Path '%s' failed at segment '%s': key not found in map\n", path, segment)
				return nil, false // Key not found
			}
			current = val
			continue
		}

		// --- CASE B: Key with Array Indices (e.g., "items[0]" or "[0]") ---

		// 1. Separate the map key from the indices
		// "items[0]" -> mapKey: "items", rest: "[0]"
		// "[0]"      -> mapKey: "",      rest: "[0]"
		mapKey := segment[:bracketIdx]
		rest := segment[bracketIdx:]

		// 2. Resolve the Map Key first (if it exists)
		if mapKey != "" {
			m, ok := current.(map[string]any)
			if !ok {
				LogMsg("Path '%s' failed at segment '%s': expected map for key '%s' but got type %T\n", path, segment, mapKey, current)
				return nil, false
			}
			val, exists := m[mapKey]
			if !exists {
				LogMsg("Path '%s' failed at segment '%s': key '%s' not found in map\n", path, segment, mapKey)
				return nil, false
			}
			current = val
		}

		// 3. Resolve Array Indices (Iterate because of cases like [0][1])
		for len(rest) > 0 {
			// Find the closing bracket
			closeIdx := strings.Index(rest, "]")
			if !strings.HasPrefix(rest, "[") || closeIdx == -1 {
				LogMsg("Path '%s' failed at array parsing: malformed brackets in '%s'\n", path, rest)
				return nil, false // Malformed path
			}

			// Parse the index number
			indexStr := rest[1:closeIdx]
			index, err := strconv.Atoi(indexStr)
			if err != nil {
				LogMsg("Path '%s' failed at array parsing: invalid index number '%s' in '%s'\n", path, indexStr, rest)
				return nil, false // Invalid number
			}

			// Assert current node is an array
			arr, ok := current.([]any)
			if !ok {
				LogMsg("Path '%s' failed: expected array at index [%d], but got type %T\n", path, index, current)
				return nil, false // Not an array
			}

			// Bounds check (Critical for stability)
			if index < 0 || index >= len(arr) {
				LogMsg("Path '%s' failed: index [%d] out of bounds (array length is %d)\n", path, index, len(arr))
				return nil, false // Index out of bounds
			}

			// Move current pointer
			current = arr[index]

			// Advance the string to check for more brackets (e.g. handle the "[1]" in "[0][1]")
			rest = rest[closeIdx+1:]
		}
	}

	return current, true
}
