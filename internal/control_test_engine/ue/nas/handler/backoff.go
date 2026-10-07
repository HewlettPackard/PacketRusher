/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 */
package handler

import (
	"my5G-RANTester/internal/control_test_engine/ue/context"
	"os"
	"time"

	"github.com/free5gc/nas/ie"
	"github.com/free5gc/nas/message"
)

// honourNetworkBackoff reports whether the UE should follow a back-off timer the network
// sent instead of its own retry schedule. Opt-in: tests that measure the simulator's own
// retry behaviour, or that compare against earlier runs, want the local schedule.
func honourNetworkBackoff() bool {
	return os.Getenv("PR_HONOUR_BACKOFF") == "1"
}

// backoffTimerFor names the timer a reject with this 5GSM cause starts. ok is false for
// the causes for which the UE shall ignore the Back-off timer value IE (TS 24.501
// 6.4.1.4.3): 28, 39, 46, 50, 51, 54, 57, 58, 61, 68 and 86. Cause 28 is among them
// because this UE's requests always carry a PDU session type IE, the condition the clause
// sets. Cause 33 is ignored only for an MA PDU session, which this UE never establishes,
// so it starts the 6.4.1.4.3 back-off timer like the other causes.
func backoffTimerFor(cause uint8) (context.BackoffTimer, bool) {
	switch cause {
	case ie.Cause5GSM_InsufRsrc:
		return context.BackoffT3396, true
	case ie.Cause5GSM_InsufRsrcForSpecificSliceAndDNN:
		return context.BackoffT3584, true
	case ie.Cause5GSM_InsufRsrcForSpecificSlice:
		return context.BackoffT3585, true
	case ie.Cause5GSM_MissingOrUnknownDNN:
		return context.BackoffDNN, true
	case ie.Cause5GSM_UnknownPDUSessType, ie.Cause5GSM_ReactivationReq, ie.Cause5GSM_OutOfLADNSvcArea,
		ie.Cause5GSM_PDUSessTypeIpv4OnlyAllowed, ie.Cause5GSM_PDUSessTypeIpv6OnlyAllowed,
		ie.Cause5GSM_PDUSessDoesNotExist, ie.Cause5GSM_PDUSessTypeIpv4V6OnlyAllowed,
		ie.Cause5GSM_PDUSessTypeUnstructuredOnlyAllowed, ie.Cause5GSM_PDUSessTypeEthOnlyAllowed,
		ie.Cause5GSM_NotSupportedSSCMode, ie.Cause5GSM_UASSvcNotAllowed:
		return 0, false
	}
	return context.BackoffDNNAndSNSSAI, true
}

// decodeBackoffTimer reads the Back-off timer value IE (TS 24.501 9.11.2.5), a GPRS
// timer 3 (TS 24.008 10.5.7.4a). ok is false when the IE cannot be read as a back-off.
//
// That is the case for the 320-hour unit, 110: table 10.5.163a NOTE 1 limits it to the
// T3312 extended value, T3412 extended value and T3512 value IEs. In any other IE it
// breaks a rule of the value part, which TS 24.007 11.4.2 defines as syntactically
// incorrect, and TS 24.501 7.7.1 has the UE treat such an optional IE as not present.
func decodeBackoffTimer(timer ie.GPRSTimer3) (wait time.Duration, deactivated, ok bool) {
	value := time.Duration(timer.Value)
	switch timer.Unit {
	case ie.TimerIncIn_2Seconds:
		return value * 2 * time.Second, false, true
	case ie.TimerIncIn_30Seconds:
		return value * 30 * time.Second, false, true
	case ie.TimerIncIn_1Minute:
		return value * time.Minute, false, true
	case ie.TimerIncIn_10Minutes:
		return value * 10 * time.Minute, false, true
	case ie.TimerIncIn_1Hour:
		return value * time.Hour, false, true
	case ie.TimerIncIn_10Hours:
		return value * 10 * time.Hour, false, true
	case ie.TimerDeactivated:
		return 0, true, true
	}
	return 0, false, false
}

// recordNetworkBackoff applies a reject's Back-off timer value IE, with
// PR_HONOUR_BACKOFF=1, to the timer its cause belongs to. Without the IE, or for a cause
// that ignores it, nothing is recorded and the local schedule applies.
func recordNetworkBackoff(ue *context.UEContext, reject *message.PDUSessEstRej) {
	if !honourNetworkBackoff() || reject.Cause5GSM == nil || reject.BackoffTimerValue == nil {
		return
	}
	timer, ok := backoffTimerFor(reject.Cause5GSM.Value)
	if !ok {
		return
	}
	wait, deactivated, ok := decodeBackoffTimer(*reject.BackoffTimerValue)
	if !ok {
		return
	}
	ue.SetEstablishmentBackoff(timer, wait, deactivated)
}
