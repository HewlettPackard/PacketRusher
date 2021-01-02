package nas

import (
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/gnb/nas/handler"
)

func Dispatch(ue *context.GNBUe, message []byte, gnb *context.GNBContext) {

	switch ue.GetState() {

	case context.Initialized:
		// handler UE message.
		handler.HandlerUeInitialized(ue, message, gnb)

	case context.Ongoing:
		// handler UE message.
		handler.HandlerUeOngoing(ue, message, gnb)

	case context.Ready:
		// handler UE message.
		handler.HandlerUeReady(ue, message, gnb)
	}
}
