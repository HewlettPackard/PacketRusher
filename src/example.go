package main

import "fmt"

func main() {

	// testing authentication for a GNB
	//fmt.Println(testAttachGnb())

	// testing attach and ping with 80 UEs.
	//fmt.Println(testMultiAttachUes(80))

	// testing UEs registration in parallel.
	fmt.Println(testMultiAttachUesInConcurrency())

}
