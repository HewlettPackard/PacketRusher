// SPDX-License-Identifier: Apache-2.0
package analytics

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestResultsCountRetriesAndCancellationSeparately(t *testing.T) {
	now := time.Unix(100, 0)
	r := newRecorder(func() time.Time { return now })
	r.Begin(1, 0, Registration)
	r.Begin(1, 0, Registration) // retransmit, not a new attempt
	now = now.Add(250 * time.Millisecond)
	r.Finish(1, 0, Registration, Success)
	r.Finish(1, 0, Registration, Success) // duplicate accept
	r.Begin(2, 0, Registration)
	r.Finish(2, 0, Registration, Failure)
	r.Begin(2, 0, Registration) // new attempt after reject
	r.Begin(1, 1, SessionEstablishment)
	r.CancelUE(1)
	r.Close()
	r.Close()
	r.Begin(3, 0, Registration) // closed reports are immutable
	report := r.Snapshot()
	require.NotNil(t, report.EndedAt)
	for _, a := range report.Procedures {
		require.Equal(t, a.Started, a.Success+a.Failure+a.Cancelled+a.Pending)
		if a.Procedure == Registration {
			require.Equal(t, uint64(3), a.Started)
			require.Equal(t, uint64(1), a.Success)
			require.Equal(t, uint64(1), a.Failure)
			require.Equal(t, uint64(1), a.Cancelled)
			require.Equal(t, .25, a.LatencyMean)
		} else {
			require.Equal(t, uint64(1), a.Cancelled)
			require.Zero(t, a.LatencyCount, "cancelled work must not skew successful latency")
		}
	}
}

func TestResultsConcurrentSnapshotsAndBoundedStorage(t *testing.T) {
	r := NewRecorder()
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(ue int64) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				r.Begin(ue, 0, Registration)
				r.Snapshot()
				r.Finish(ue, 0, Registration, Success)
			}
		}(int64(i))
	}
	wg.Wait()
	require.Empty(t, r.pending)
	require.Len(t, r.totals, 2)
	for _, a := range r.Snapshot().Procedures {
		if a.Procedure == Registration {
			require.Equal(t, uint64(2000), a.Success)
		}
	}
}

func TestResultsExportsAndPrometheusHistogram(t *testing.T) {
	now := time.Unix(100, 0)
	r := newRecorder(func() time.Time { return now })
	r.Begin(999999, 0, Registration)
	now = now.Add(10 * time.Millisecond)
	r.Finish(999999, 0, Registration, Success)
	r.Close()
	var b bytes.Buffer
	require.NoError(t, r.WriteJSON(&b))
	var report Report
	require.NoError(t, json.Unmarshal(b.Bytes(), &report))
	require.Equal(t, 1, report.SchemaVersion)
	require.NotContains(t, b.String(), "999999", "UE identity must not be exported")
	b.Reset()
	require.NoError(t, r.WriteCSV(&b))
	rows, err := csv.NewReader(&b).ReadAll()
	require.NoError(t, err)
	require.Len(t, rows, 3)
	require.Equal(t, "procedure", rows[0][0])
	response := httptest.NewRecorder()
	r.ServeHTTP(response, httptest.NewRequest("GET", "/metrics", nil))
	require.Equal(t, 200, response.Code)
	body := response.Body.String()
	require.Contains(t, body, `packetrusher_procedure_completed_total{procedure="registration",outcome="success"} 1`)
	require.Contains(t, body, `packetrusher_procedure_duration_seconds_bucket{procedure="registration",le="0.005"} 0`)
	require.Contains(t, body, `packetrusher_procedure_duration_seconds_bucket{procedure="registration",le="0.01"} 1`)
	require.Contains(t, body, `packetrusher_procedure_duration_seconds_bucket{procedure="registration",le="+Inf"} 1`)
	require.NotContains(t, body, "999999")
	require.True(t, strings.HasSuffix(body, "\n"))
	missing := httptest.NewRecorder()
	r.ServeHTTP(missing, httptest.NewRequest("GET", "/", nil))
	require.Equal(t, 404, missing.Code)
}

func TestReportTimestampsUseUTC(t *testing.T) {
	local := time.Date(2026, 10, 2, 18, 30, 0, 0, time.FixedZone("test-local", 2*60*60))
	r := newRecorder(func() time.Time { return local })
	if report := r.Snapshot(); report.StartedAt.Location() != time.UTC || !report.StartedAt.Equal(local) {
		t.Fatalf("start=%v; want the same instant in UTC", report.StartedAt)
	}
	local = local.Add(time.Minute)
	r.Close()
	report := r.Snapshot()
	if report.EndedAt == nil || report.EndedAt.Location() != time.UTC || !report.EndedAt.Equal(local) {
		t.Fatalf("end=%v; want the same instant in UTC", report.EndedAt)
	}
	var output bytes.Buffer
	require.NoError(t, r.WriteJSON(&output))
	require.Contains(t, output.String(), `"started_at": "2026-10-02T16:30:00Z"`)
	require.Contains(t, output.String(), `"ended_at": "2026-10-02T16:31:00Z"`)
}
