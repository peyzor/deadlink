package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("incorrect number of given arguments expected 1 but got: %d", len(os.Args)-1)
		return
	}

	link := os.Args[1]
	fmt.Println(link)
}
