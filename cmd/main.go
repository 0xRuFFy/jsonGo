package main

import (
	"fmt"

	jsongo "github.com/0xRuFFy/jsonGo"
)

const FILE = "./testdata/level_3.json"

func main() {
	tokenizer, err := jsongo.NewJsonTokenizer(FILE)
	if err != nil {
		panic(err)
	}
	fmt.Println(tokenizer.FileContent)

	// token, err := tokenizer.NextToken()
	// if err != nil {
	// 	panic(err)
	// }

	// for token.Type != jsongo.JTT_EOF {
	// 	fmt.Println(token.String())
	// 	token, err = tokenizer.NextToken()
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }

	json, err := jsongo.ParseFile(FILE)
	if err != nil {
		panic(err)
	}

	fmt.Println(json.Data)
}
