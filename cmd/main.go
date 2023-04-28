package main

import (
	"fmt"

	jsongo "github.com/0xRuFFy/jsonGo"
)

const FILE = "./testdata/level_3.json"

func main() {
	json, err := jsongo.ParseFile(FILE)
	if err != nil {
		panic(err)
	}

	fmt.Println(json.Data)
}
