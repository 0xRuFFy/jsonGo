package main

import (
	"fmt"

	jsongo "github.com/0xRuFFy/jsonGo"
)

const FILE = "./testdata/canada.json"

func main() {
	json, err := jsongo.ParseFile(FILE)
	if err != nil {
		panic(err)
	}

	switch json.Data["main"].(type) {
	case []interface{}:
		fmt.Println("main is array")
		fmt.Println(len(json.Data["main"].([]interface{})))
	case map[string]interface{}:
		fmt.Println("main is map")
	default:
		fmt.Println("main is unknown")
	}
	// fmt.Println(json.Data)
}
