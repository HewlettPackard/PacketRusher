package main

import "fmt"

func main() {

	// testing authentication for a GNB
	//fmt.Println(testAttachGnb())

	// testing attach and ping with 80 UEs.
	// fmt.Println(testMultiAttachUesInQueue(1))

	// testing concurrent UEs registration with GNBs.
	// fmt.Println(testMultiAttachUesInConcurrencyWithGNBs())

	// testing concurrent UEs registration with TNLAs.
	fmt.Println(testMultiAttachUesInConcurrencyWithTNLAs(10))

	// testing multiple GNBs authentication(control plane only)-> NGAP Request and response tester.
	// fmt.Println(testMultiAttachGnbInQueue(100))
	// fmt.Println(testMultiAttachGnbInConcurrency(200))
}
