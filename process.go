package main

import (
	"fmt"
	"strings"
)

// processHeader iterates through the provided header map and performs variable substitution
// on all values. It modifies the map in-place.
// Returns false if any variable substitution fails.
func processHeader(header map[string]string) bool {
	for k, v := range header {
		// Attempt to substitute variables in the header value
		temp, ok := processString(v)
		if !ok {
			return false
		}
		header[k] = temp
	}
	return true
}

// processUrl performs variable substitution on the request URL.
// Returns the processed URL and a boolean indicating success.
func processUrl(url string) (string, bool) {
	return processString(url)
}

// processBody determines the underlying type of the body (map or slice)
// and delegates to the appropriate processing function.
// It supports dynamic JSON structures deserialized into 'any'.
func processBody(data any) bool {
	current, ok := data.(map[string]any)
	if !ok {
		// If not a map, check if it is a slice/array
		current2, ok := data.([]any)
		if !ok {
			fmt.Println("Given data doesn't seem to be either array or object")
			return false
		} else {
			return processArray(current2)
		}
	} else {
		return processMap(current)
	}
}

// processMap recursively traverses a map to find and substitute strings.
// It handles nested maps and arrays within the map.
func processMap(current map[string]any) bool {
	var ok bool
	for key, val := range current {
		switch v := val.(type) {
		case []any:
			// Recursively process nested arrays
			processArray(v)
		case map[string]any:
			// Recursively process nested maps
			processMap(v)
		case string:
			// Perform substitution on string values
			current[key], ok = processString(v)
			if !ok {
				return false
			}
		}
	}
	return true
}

// processArray recursively traverses a slice to find and substitute strings.
// It handles nested maps and arrays within the slice.
func processArray(v []any) bool {
	var ok bool
	for arrI, arrItem := range v {
		switch arrV := arrItem.(type) {
		case []any:
			// Recursively process nested arrays
			processArray(arrV)
		case map[string]any:
			// Recursively process nested maps
			processMap(arrV)
		case string:
			// Perform substitution on string elements
			v[arrI], ok = processString(arrV)
			if !ok {
				return false
			}
		}
	}
	return true
}

// processString parses a string to identify and replace variable placeholders.
// It expects variables to be delimited by '$' (e.g., $VAR_NAME$).
// It looks up values in the global 'variables' map.
func processString(str string) (string, bool) {
	result := ""
	firstPassed := false
	varName := ""
	// Iterate over the string using a sequence (Go 1.23+)
	for v := range strings.SplitSeq(str, "") {
		if v != "$" && !firstPassed {
			// standard character, append to result
			result += v
		} else if v == "$" && !firstPassed {
			// start of a variable declaration
			firstPassed = true
		} else if v != "$" && firstPassed {
			// accumulation of the variable name
			varName += v
		} else if v == "$" && firstPassed {
			// end of variable declaration, perform lookup
			firstPassed = false
			// result += variables[varName]
			t, ok := variables[varName]
			if !ok {
				fmt.Printf("%v is not present in variables.\n", varName)
				return "", false
			}
			temp, ok := t.(string)
			if !ok {
				fmt.Printf("Couldn't convert variable %v to string.\n", varName)
				return "", false
			}
			result += temp
			varName = ""
		}
	}
	return result, true
}
