/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 */
package handler

import (
	"math"
	"os"
	"time"

	"github.com/free5gc/nas/nasMessage"
)

// GPRS timer 3 units that free5gc defines no constant for (TS 24.008 10.5.7.4a).
const (
	gprsTimer3UnitMultiplesOf320Hours uint8 = 0x06
	gprsTimer3UnitDeactivated         uint8 = 0x07
)

// 5GSM causes that free5gc defines no constant for (TS 24.501 9.11.4.2).
const (
	cause5GSMPDUSessionTypeIPv4v6OnlyAllowed       uint8 = 57
	cause5GSMPDUSessionTypeUnstructuredOnlyAllowed uint8 = 58
	cause5GSMPDUSessionTypeEthernetOnlyAllowed     uint8 = 61
	cause5GSMUASServicesNotAllowed                 uint8 = 86
)

// backoffIgnoredFor holds the 5GSM causes for which the UE shall ignore the Back-off
// timer value IE (TS 24.501 6.4.1.4.3). #28 is listed because this UE's requests always
// carry a PDU session type IE, the condition the clause sets for it. #33 is left out: it
// applies only to an MA PDU session, which this UE never establishes.
var backoffIgnoredFor = map[uint8]bool{
	nasMessage.Cause5GSMUnknownPDUSessionType:         true,
	nasMessage.Cause5GSMReactivationRequested:         true,
	nasMessage.Cause5GSMOutOfLADNServiceArea:          true,
	nasMessage.Cause5GSMPDUSessionTypeIPv4OnlyAllowed: true,
	nasMessage.Cause5GSMPDUSessionTypeIPv6OnlyAllowed: true,
	cause5GSMPDUSessionTypeIPv4v6OnlyAllowed:          true,
	cause5GSMPDUSessionTypeUnstructuredOnlyAllowed:    true,
	cause5GSMPDUSessionTypeEthernetOnlyAllowed:        true,
	nasMessage.Cause5GSMPDUSessionDoesNotExist:        true,
	nasMessage.Cause5GSMNotSupportedSSCMode:           true,
	cause5GSMUASServicesNotAllowed:                    true,
}

// honourNetworkBackoff reports whether the UE should follow a back-off timer the
// network sent instead of its own retry schedule. Opt-in: tests that measure the
// simulator's own retry behaviour, or that compare against earlier runs, want the
// local schedule left alone.
func honourNetworkBackoff() bool {
	return os.Getenv("PR_HONOUR_BACKOFF") == "1"
}

// networkBackoff reads the Back-off timer value IE a PDU Session Establishment Reject
// may carry (TS 24.501 8.3.3.2), a GPRS timer 3 (TS 24.008 10.5.7.4a). TS 24.501
// 6.4.1.4.2 and 6.4.1.4.3 give the UE three cases:
//   - a value that is neither zero nor deactivated: no further request before it ends;
//   - zero: the UE may send another request straight away;
//   - deactivated: the UE shall not send another request for the same DNN.
//
// ok is false when the reject carries no such IE. When ok is true, retry says whether
// another request may be sent at all, and wait how long before it.
func networkBackoff(reject *nasMessage.PDUSessionEstablishmentReject) (wait time.Duration, retry, ok bool) {
	if reject == nil || reject.BackoffTimerValue == nil {
		return 0, false, false
	}
	value := time.Duration(reject.BackoffTimerValue.GetTimerValue())
	switch reject.BackoffTimerValue.GetUnitTimerValue() {
	case nasMessage.GPRSTimer3UnitMultiplesOf2Seconds:
		return value * 2 * time.Second, true, true
	case nasMessage.GPRSTimer3UnitMultiplesOf30Seconds:
		return value * 30 * time.Second, true, true
	case nasMessage.GPRSTimer3UnitMultiplesOf1Minute:
		return value * time.Minute, true, true
	case nasMessage.GPRSTimer3UnitMultiplesOf10Minutes:
		return value * 10 * time.Minute, true, true
	case nasMessage.GPRSTimer3UnitMultiplesOf1Hour:
		return value * time.Hour, true, true
	case nasMessage.GPRSTimer3UnitMultiplesOf10Hours:
		return value * 10 * time.Hour, true, true
	case gprsTimer3UnitMultiplesOf320Hours:
		return value * 320 * time.Hour, true, true
	}
	// The only unit left in three bits is "deactivated".
	return 0, false, true
}

// retryDecision is what a PDU Session Establishment Reject leaves the UE to do.
type retryDecision struct {
	wait        time.Duration // before the next request
	retry       bool          // whether to send another request at all
	fromNetwork bool          // whether the network's back-off timer set the above
}

// retryAfterReject decides whether, and after how long, the UE requests a rejected PDU
// session again, given how many retries it has made. By default it keeps the simulator's
// own schedule, 5^retries seconds. With PR_HONOUR_BACKOFF=1 it follows the reject's
// Back-off timer value IE, unless the IE is absent or the cause is one for which the UE
// shall ignore it (TS 24.501 6.4.1.4.3).
func retryAfterReject(reject *nasMessage.PDUSessionEstablishmentReject, retries int) retryDecision {
	local := retryDecision{wait: time.Duration(math.Pow(5, float64(retries))) * time.Second, retry: true}
	if !honourNetworkBackoff() || reject == nil || backoffIgnoredFor[reject.GetCauseValue()] {
		return local
	}
	wait, retry, ok := networkBackoff(reject)
	if !ok {
		return local
	}
	return retryDecision{wait: wait, retry: retry, fromNetwork: true}
}
