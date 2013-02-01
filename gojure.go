package main

import (
	"./edn"
	"fmt"
)

func main() {
	for {
		fmt.Print(">> ")

		var input string
		fmt.Scanf("%s", &input)
		edn.Read(&input)
	}
}
