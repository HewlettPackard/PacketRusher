package logging

import "fmt"

func Check_error(err error, message string) bool {

	if err != nil {
		return true
	} else {
		fmt.Println(message)
		return false
	}
}
