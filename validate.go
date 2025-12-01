package main

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// validateBody checks if actual matches the structure/values of expected.
// It implements "Expected <= Actual" logic (Subset Validation).
// Now supports Regex strings and Unordered Array Subset Matching.
func validateBody(expected, actual any) bool {
	if expected == nil {
		return true
	}

	switch exp := expected.(type) {
	case map[string]any:
		// Actual must be a map
		act, ok := actual.(map[string]any)
		if !ok {
			return false
		}
		// Every key in Expected must exist in Actual and match value recursively
		for k, vExp := range exp {
			vAct, exists := act[k]
			if !exists {
				return false // Key missing in actual
			}
			if !validateBody(vExp, vAct) {
				return false // Value mismatch
			}
		}
		return true

	case []any:
		// Actual must be an array
		act, ok := actual.([]any)
		if !ok {
			return false
		}
		// Basic length check: Actual must contain at least as many items as Expected
		if len(act) < len(exp) {
			return false
		}

		// Unordered Subset Match Strategy:
		// For each item in 'expected', we must find a unique matching item in 'actual' (anywhere).
		// We use matchedIndices to ensure we don't match the same actual item twice.
		matchedIndices := make([]bool, len(act))

		for _, expItem := range exp {
			found := false
			// Scan the entire actual array for a match
			for j, actItem := range act {
				if matchedIndices[j] {
					continue // This actual item is already "consumed" by a previous expected item
				}
				if validateBody(expItem, actItem) {
					matchedIndices[j] = true
					found = true
					break // Match found for this expected item, move to next
				}
			}
			// If we iterated through all actual items and found no match for this expected item
			if !found {
				return false
			}
		}
		return true

	case string:
		// Regular String & Regex Support
		actStr, ok := actual.(string)
		if !ok {
			return false
		}
		if strings.HasPrefix(exp, "regex:") {
			pattern := strings.TrimPrefix(exp, "regex:")
			matched, err := regexp.MatchString(pattern, actStr)
			if err != nil {
				fmt.Printf("Invalid regex pattern: %s. Error: %v\n", pattern, err)
				return false
			}
			return matched
		}
		return exp == actStr

	default:
		// Primitives (float64, bool, etc.) must match exactly
		return reflect.DeepEqual(expected, actual)
	}
}

// Based on the `validateBody` function in your **Canvas** file, here is a breakdown of its capabilities and limitations:
//
// ###  What it CAN Do
//
// 1.  **Partial / Subset Object Matching:**
//     * It implements **"Expected $\subseteq$ Actual"**.
//     * You only need to define the fields you care about in `expected_response`.
//     * Example: If `Actual` is `{"id": 1, "name": "foo", "date": "..."}` and `Expected` is `{"name": "foo"}`, it **PASSES**.
//
// 2.  **Unordered Array Matching:**
//     * It checks if items exist *anywhere* in the array. Order does not matter.
//     * Example: If `Expected` is `[1, 2]` and `Actual` is `[3, 2, 1]`, it **PASSES**.
//     * It handles duplicates correctly (e.g., `Expected: [1, 1]` requires two `1`s in `Actual`).
//
// 3.  **Pattern Matching (Regex):**
//     * It supports dynamic string validation using the `regex:` prefix.
//     * Example: `Expected: "regex:^user_\\d+$"` matches `Actual: "user_123"`.
//
// 4.  **Deep Recursive Validation:**
//     * It works on deeply nested structures (e.g., an Object inside an Array inside an Object).
//
// 5.  **Ignore Extra Data:**
//     * It ignores extra fields in objects and extra items in arrays (as long as the expected ones are found).
//
// ---
//
// ###  What it CANNOT Do
//
// 1.  **Enforce Strict Order in Arrays:**
//     * You cannot force `[A, B]` to appear exactly in that order. `[B, A]` will pass.
//     * *Limitation:* If order matters (e.g., a sorted list API), this validation logic is too loose.
//
// 2.  **Enforce "Exact Match" (No Extra Fields):**
//     * You cannot ensure that the response *only* contains the fields you specified.
//     * If the API leaks sensitive data (e.g., `password_hash`) that wasn't in your `Expected` JSON, the test will still pass.
//
// 3.  **Negative Assertions:**
//     * You cannot explicitly check that a field **does not exist** (e.g., ensuring `error` field is missing).
//
// 4.  **Type Coercion:**
//     * It uses `reflect.DeepEqual` for primitives.
//     * `1` (integer) does not equal `1.0` (float).
//     * *Note:* Since both `Expected` and `Actual` come from `json.Unmarshal`, Go treats all numbers as `float64` by default, so this is rarely an issue unless you manually construct the expected data in Go code with specific types.
//
// 5.  **Complex Logic:**
//     * It cannot perform logic like "Value A must be greater than Value B" or "Length of array must be exactly 5". It checks only for existence and equality/regex match.
