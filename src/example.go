package main

import "fmt"

func main() {

	// testing authentication for a GNB
	//fmt.Println(testAttachGnb())

	// testing attach and ping with 80 UEs.
	// fmt.Println(testMultiAttachUesInQueue(80))

	// testing concurrent UEs registration with GNBs.
	// fmt.Println(testMultiAttachUesInConcurrencyWithGNBs())

	// testing concurrent UEs registration with TNLAs.
	fmt.Println(testMultiAttachUesInConcurrencyWithTNLAs(2))
}
