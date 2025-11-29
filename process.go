package main

import (
	"fmt"
	"strings"
)

func processHeader(header map[string]string) bool {
	for k, v := range header {
		temp, ok := processString(v)
		if !ok {
			return false
		}
		header[k] = temp
	}
	return true
}

func processString(str string) (string, bool) {
	result := ""
	firstPassed := false
	varName := ""
	for v := range strings.SplitSeq(str, "") {
		if v != "$" && !firstPassed {
			result += v
		} else if v == "$" && !firstPassed {
			firstPassed = true
		} else if v != "$" && firstPassed {
			varName += v
		} else if v == "$" && firstPassed {
			firstPassed = false
			// result += variables[varName]
			temp, ok := variables[varName].(string)
			if !ok {
				fmt.Printf("Error in converting variable %v to string.\n", varName)
				return "", false
			}
			result += temp
			varName = ""
		}
	}
	return result, true
}
