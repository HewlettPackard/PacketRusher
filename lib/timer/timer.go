package timer

import (
	"log"
	"time"
)

func StartTimer(seconds int, action func(interface{}), msg interface{}) *time.Timer {
	timer := time.NewTimer(time.Second * time.Duration(seconds))

	go func() {
		select {
		case <-timer.C:
			action(msg)
		case <-time.After(time.Second*time.Duration(seconds) + 100*time.Millisecond):
			log.Println("timer closed")
		}
	}()

	return timer
}
