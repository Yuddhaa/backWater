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

func processBody(data any) bool {
	current, ok := data.(map[string]any)
	if !ok {
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

func processMap(current map[string]any) bool {
	var ok bool
	for key, val := range current {
		switch v := val.(type) {
		case []any:
			processArray(v)
		case map[string]any:
			processMap(v)
		case string:
			current[key], ok = processString(v)
			if !ok {
				return false
			}
		}
	}
	return true
}

func processArray(v []any) bool {
	var ok bool
	for arrI, arrItem := range v {
		switch arrV := arrItem.(type) {
		case []any:
			processArray(arrV)
		case map[string]any:
			processMap(arrV)
		case string:
			v[arrI], ok = processString(arrV)
			if !ok {
				return false
			}
		}
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
