/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */
package analytics

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestResultsAsJSONAndPrometheus(t *testing.T) {
	var registration, session time.Time
	Start(Registration, &registration)
	Start(Registration, &registration) // a retransmission is the same attempt
	registration = registration.Add(-3 * time.Second)
	Finish(Registration, &registration, true)
	Finish(Registration, &registration, true) // nothing pending: not counted
	Start(Registration, &registration)
	Finish(Registration, &registration, false)
	Start(SessionEstablishment, &session)

	var output bytes.Buffer
	require.NoError(t, WriteJSON(&output))
	var report struct {
		StartedAt  time.Time `json:"started_at"`
		EndedAt    time.Time `json:"ended_at"`
		Procedures []map[string]any
	}
	require.NoError(t, json.Unmarshal(output.Bytes(), &report))
	require.False(t, report.EndedAt.Before(report.StartedAt))
	require.Len(t, report.Procedures, 2)
	latency := report.Procedures[0]["latency_seconds_sum"]
	require.InDelta(t, 3, latency, 1)
	require.Equal(t, map[string]any{
		"procedure": "registration", "started": 2.0, "success": 1.0, "failure": 1.0, "pending": 0.0, "latency_count": 1.0,
		"latency_seconds_sum": latency, "latency_seconds_min": latency, "latency_seconds_max": latency, "latency_seconds_mean": latency,
	}, report.Procedures[0])
	require.Equal(t, map[string]any{
		"procedure": "pdu_session_establishment", "started": 1.0, "success": 0.0, "failure": 0.0, "pending": 1.0, "latency_count": 0.0,
		"latency_seconds_sum": 0.0, "latency_seconds_min": 0.0, "latency_seconds_max": 0.0, "latency_seconds_mean": 0.0,
	}, report.Procedures[1])

	response := httptest.NewRecorder()
	Metrics(response, httptest.NewRequest("GET", "/metrics", nil))
	for _, line := range []string{
		"# TYPE packetrusher_procedure_duration_seconds histogram",
		`packetrusher_procedure_started_total{procedure="registration"} 2`,
		`packetrusher_procedure_completed_total{procedure="registration",outcome="success"} 1`,
		`packetrusher_procedure_completed_total{procedure="registration",outcome="failure"} 1`,
		`packetrusher_procedure_pending{procedure="pdu_session_establishment"} 1`,
		`packetrusher_procedure_duration_seconds_bucket{procedure="registration",le="2.5"} 0`,
		`packetrusher_procedure_duration_seconds_bucket{procedure="registration",le="5"} 1`,
		`packetrusher_procedure_duration_seconds_bucket{procedure="registration",le="+Inf"} 1`,
		`packetrusher_procedure_duration_seconds_count{procedure="registration"} 1`,
	} {
		require.Contains(t, response.Body.String(), line+"\n")
	}
}
