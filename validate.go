package main

import (
	"reflect"
	"regexp"
	"strings"
)

// validateBody checks if actual matches expected (Subset + Regex + Unordered Array).
// It now accepts 'quiet' bool. If true, it suppress logs (useful for speculative matching in arrays).
// If false, it uses LogMsg to report specific mismatches.
func validateBody(expected, actual any, quiet bool) bool {
	if expected == nil {
		return true
	}

	switch exp := expected.(type) {
	case map[string]any:
		act, ok := actual.(map[string]any)
		if !ok {
			if !quiet {
				LogMsg("[Validation Error] Expected JSON Object, got %T\n", actual)
			}
			return false
		}
		for k, vExp := range exp {
			vAct, exists := act[k]
			if !exists {
				if !quiet {
					LogMsg("[Validation Error] Missing expected key: '%s'\n", k)
				}
				return false
			}
			if !validateBody(vExp, vAct, quiet) {
				if !quiet {
					LogMsg("[Validation Error] Mismatch at key: '%s'\n", k)
				}
				return false
			}
		}
		return true

	case []any:
		act, ok := actual.([]any)
		if !ok {
			if !quiet {
				LogMsg("[Validation Error] Expected JSON Array, got %T\n", actual)
			}
			return false
		}
		if len(act) < len(exp) {
			if !quiet {
				LogMsg("[Validation Error] Actual array length (%d) is less than expected (%d)\n", len(act), len(exp))
			}
			return false
		}
		matchedIndices := make([]bool, len(act))
		for _, expItem := range exp {
			found := false
			for j, actItem := range act {
				if matchedIndices[j] {
					continue
				}
				// Try match silently first
				if validateBody(expItem, actItem, true) {
					matchedIndices[j] = true
					found = true
					break
				}
			}
			if !found {
				if !quiet {
					LogMsg("[Validation Error] Could not find match for expected item: %v\n", expItem)
				}
				return false
			}
		}
		return true

	case string:
		actStr, ok := actual.(string)
		if !ok {
			if !quiet {
				LogMsg("[Validation Error] Expected String, got %T\n", actual)
			}
			return false
		}
		if strings.HasPrefix(exp, "regex:") {
			pattern := strings.TrimPrefix(exp, "regex:")
			matched, err := regexp.MatchString(pattern, actStr)
			if err != nil {
				if !quiet {
					LogMsg("[Validation Error] Invalid regex pattern '%s': %v\n", pattern, err)
				}
				return false
			}
			if !matched && !quiet {
				LogMsg("[Validation Error] Value '%s' did not match regex '%s'\n", actStr, pattern)
			}
			return matched
		}
		if exp != actStr && !quiet {
			LogMsg("[Validation Error] Expected string '%s', got '%s'\n", exp, actStr)
		}
		return exp == actStr

	default:
		match := reflect.DeepEqual(expected, actual)
		if !match && !quiet {
			LogMsg("[Validation Error] Value mismatch. Expected %v (%T), Got %v (%T)\n", expected, expected, actual, actual)
		}
		return match
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
