package main

import (
	"fmt"
)

func main() {

	// testing attach and ping with a number of UEs.
	// fmt.Println( templates.TestMultiAttachUesInQueue(1) )

	// testing concurrent UEs registration with GNBs.
	// fmt.Println( templates.TestMultiAttachUesInConcurrencyWithGNBs() )

	// testing concurrent UEs registration with some SCTPs associations.
	// fmt.Println( templates.TestMultiAttachUesInConcurrencyWithTNLAs(1) )

	// testing multiple GNBs authentication(control plane only)-> NGAP Request and response tester.
	fmt.Println(testMultiAttachGnbInQueue(100))
	// fmt.Println( templates.TestMultiAttachGnbInConcurrency(1) )
}
