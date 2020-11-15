package logging

import (
	"fmt"
	"my5G-RANTester/lib/ngap/ngapType"
)

func Check_error(err error, message string) bool {

	if err != nil {
		return true
	} else {
		fmt.Println(message)
		return false
	}
}

func Check_Ngap(ngap *ngapType.NGAPPDU, message string) bool {

	if ngap == nil {
		return true
	} else {
		fmt.Println(message)
		return false
	}
}
