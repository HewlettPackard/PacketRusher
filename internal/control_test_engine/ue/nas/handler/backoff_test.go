/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Forsway Scandinavia AB
 */
package handler

import (
	"testing"
	"time"

	"github.com/free5gc/nas/nasMessage"
	"github.com/free5gc/nas/nasType"
	"github.com/stretchr/testify/assert"
)

func withBackoff(r *nasMessage.PDUSessionEstablishmentReject, unit, value uint8) *nasMessage.PDUSessionEstablishmentReject {
	r.BackoffTimerValue = nasType.NewBackoffTimerValue(nasMessage.PDUSessionEstablishmentRejectBackoffTimerValueType)
	r.BackoffTimerValue.SetLen(1)
	r.BackoffTimerValue.SetUnitTimerValue(unit)
	r.BackoffTimerValue.SetTimerValue(value)
	return r
}

func TestNetworkBackoffDecodesEveryUnit(t *testing.T) {
	for _, c := range []struct {
		unit uint8
		want time.Duration
	}{
		{nasMessage.GPRSTimer3UnitMultiplesOf2Seconds, 6 * time.Second},
		{nasMessage.GPRSTimer3UnitMultiplesOf30Seconds, 90 * time.Second},
		{nasMessage.GPRSTimer3UnitMultiplesOf1Minute, 3 * time.Minute},
		{nasMessage.GPRSTimer3UnitMultiplesOf10Minutes, 30 * time.Minute},
		{nasMessage.GPRSTimer3UnitMultiplesOf1Hour, 3 * time.Hour},
		{nasMessage.GPRSTimer3UnitMultiplesOf10Hours, 30 * time.Hour},
		{gprsTimer3UnitMultiplesOf320Hours, 960 * time.Hour},
	} {
		wait, retry, ok := networkBackoff(withBackoff(reject(1), c.unit, 3))
		assert.True(t, ok, "unit %d", c.unit)
		assert.True(t, retry, "unit %d", c.unit)
		assert.Equal(t, c.want, wait, "unit %d", c.unit)
	}
}

func TestRetryAfterReject(t *testing.T) {
	r := withBackoff(reject(1), nasMessage.GPRSTimer3UnitMultiplesOf2Seconds, 4)

	t.Setenv("PR_HONOUR_BACKOFF", "")
	assert.Equal(t, retryDecision{wait: 25 * time.Second, retry: true}, retryAfterReject(r, 2),
		"off by default: the local 5^n schedule")

	t.Setenv("PR_HONOUR_BACKOFF", "true")
	assert.Equal(t, retryDecision{wait: 25 * time.Second, retry: true}, retryAfterReject(r, 2),
		"only 1 turns it on, as for PR_VERIFY_RULES")

	t.Setenv("PR_HONOUR_BACKOFF", "1")
	assert.Equal(t, retryDecision{wait: 8 * time.Second, retry: true, fromNetwork: true}, retryAfterReject(r, 2),
		"honoured: the network's value")
	assert.Equal(t, retryDecision{wait: 5 * time.Second, retry: true}, retryAfterReject(reject(1), 1),
		"no IE: the local schedule")
	assert.Equal(t, retryDecision{retry: false, fromNetwork: true},
		retryAfterReject(withBackoff(reject(1), gprsTimer3UnitDeactivated, 1), 0), "deactivated: no retry")
}

// TS 24.501 6.4.1.4.3: for these causes the UE shall ignore the Back-off timer value IE.
func TestRetryAfterRejectIgnoresTheIEForSomeCauses(t *testing.T) {
	t.Setenv("PR_HONOUR_BACKOFF", "1")
	for _, cause := range []uint8{28, 39, 46, 50, 51, 54, 57, 58, 61, 68, 86} {
		r := withBackoff(reject(1), gprsTimer3UnitDeactivated, 0)
		r.SetCauseValue(cause)
		assert.Equal(t, retryDecision{wait: time.Second, retry: true}, retryAfterReject(r, 0),
			"cause %d: the IE is ignored and the local schedule applies", cause)
	}
}
