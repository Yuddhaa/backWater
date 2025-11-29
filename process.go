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

func processUrl(url string) (string, bool) {
	return processString(url)
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
			t, ok := variables[varName]
			if !ok {
				fmt.Printf("%v is not present in variables.\n", varName)
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
