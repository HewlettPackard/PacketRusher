package state

import (
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"my5G-RANTester/internal/control_test_engine/ue/nas"
)

func DispatchState(ue *context.UEContext, message []byte) {
	nas.DispatchNas(ue, message)
}
