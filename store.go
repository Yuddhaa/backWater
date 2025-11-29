package main

import (
	"encoding/json"
	"fmt"
	"maps"
	"strconv"
	"strings"
)

func storeGlobalVariables(variables, input variablesStruct) {
	maps.Copy(variables, input)
}

func storeBodyVariables(testNo int, body []byte, variables variablesStruct, toStore map[string]string) bool {
	if len(body) == 0 || body[0] != '{' && body[0] != '[' {
		fmt.Println("Body is not JSON, skipping unmarshal.")
		return true
	}
	// fmt.Println("body:", string(body))
	var tempBody map[string]any
	if err := json.Unmarshal(body, &tempBody); err != nil {
		fmt.Println("Error in Unmarshal of the body. Err:" + err.Error())
		return false
	}
	for k, v := range toStore {
		n := strconv.FormatInt(int64(testNo), 10)
		varValue, ok := getNestedValue(v, tempBody)
		if !ok {
			return false
		}
		variables["test_"+n+"_"+k] = varValue
	}
	return true
}

func getNestedValue(path string, data any) (any, bool) {
	keys := strings.Split(path, ".")
	current := data
	for _, key := range keys {
		iIdx := strings.Index(key, "[") + 1
		var i int
		var err error
		isArr := false
		if iIdx > 0 {
			i, err = strconv.Atoi(string(key[iIdx]))
			if err != nil {
				fmt.Printf("Invalid index in path: %v. Error:%v\n", path, err.Error())
				return nil, false
			}
			key = strings.Split(key, "[")[0]
			isArr = true
		}

		m, mok := current.(map[string]any)
		if !mok {
			fmt.Printf("path segment '%s' not found or parent is not a map\n", key)
			return nil, false
		}
		val, exists := m[key]
		if !exists {
			fmt.Printf("key '%s' not found\n", key)
			return nil, false
		}
		if isArr {
			fmt.Println("inside is arr if")
			arr, ok := val.([]any)
			if !ok {
				fmt.Printf("path segment '%s' is not an array\n", key)
				return nil, false
			}
			current = arr[i]
		} else {
			current = val
		}
	}
	return current, true
}
