package main

import "maps"

func storeGlobalVariables(variables, input variablesStruct) {
	maps.Copy(variables, input)
}
