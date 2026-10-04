/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */

// Package analytics counts the registrations and PDU session establishments of
// the UEs, and reports them as JSON and as Prometheus metrics.
package analytics

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type Procedure int

const (
	Registration Procedure = iota
	SessionEstablishment
)

// latencyBounds are the upper bounds, in seconds, of the latency histogram.
var latencyBounds = [...]float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}

// Result counts the attempts of one procedure. An attempt neither succeeded nor
// failed is pending. The latencies are those of the successful attempts.
type Result struct {
	Procedure    string  `json:"procedure"`
	Started      uint64  `json:"started"`
	Success      uint64  `json:"success"`
	Failure      uint64  `json:"failure"`
	Pending      uint64  `json:"pending"`
	LatencyCount uint64  `json:"latency_count"`
	LatencySum   float64 `json:"latency_seconds_sum"`
	LatencyMin   float64 `json:"latency_seconds_min"`
	LatencyMax   float64 `json:"latency_seconds_max"`
	LatencyMean  float64 `json:"latency_seconds_mean"`
	buckets      [len(latencyBounds)]uint64
}

var (
	mu        sync.Mutex
	startedAt = time.Now().UTC()
	results   = [...]Result{{Procedure: "registration"}, {Procedure: "pdu_session_establishment"}}
)

// Start counts a new attempt and records in start, kept by the UE or its PDU
// session, when it began. A retransmission of a pending attempt is not counted.
func Start(procedure Procedure, start *time.Time) {
	if !start.IsZero() {
		return
	}
	*start = time.Now()
	mu.Lock()
	results[procedure].Started++
	mu.Unlock()
}

// Finish counts the outcome of the attempt that began at start, if there is one.
func Finish(procedure Procedure, start *time.Time, success bool) {
	if start.IsZero() {
		return
	}
	seconds := time.Since(*start).Seconds()
	*start = time.Time{}
	mu.Lock()
	defer mu.Unlock()
	r := &results[procedure]
	if !success {
		r.Failure++
		return
	}
	r.Success++
	r.LatencySum += seconds
	if r.Success == 1 || seconds < r.LatencyMin {
		r.LatencyMin = seconds
	}
	r.LatencyMax = max(r.LatencyMax, seconds)
	for i, bound := range latencyBounds {
		if seconds <= bound {
			r.buckets[i]++
		}
	}
}

// Results returns the current counters of every procedure.
func Results() []Result {
	mu.Lock()
	snapshot := results
	mu.Unlock()
	for i := range snapshot {
		r := &snapshot[i]
		r.Pending = r.Started - r.Success - r.Failure
		r.LatencyCount = r.Success
		if r.Success != 0 {
			r.LatencyMean = r.LatencySum / float64(r.Success)
		}
	}
	return snapshot[:]
}

// WriteJSON writes the report of the run.
func WriteJSON(w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(struct {
		StartedAt  time.Time `json:"started_at"`
		EndedAt    time.Time `json:"ended_at"`
		Procedures []Result  `json:"procedures"`
	}{startedAt, time.Now().UTC(), Results()})
}

// Metrics serves the counters in the Prometheus text format.
func Metrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	results := Results()
	fmt.Fprintln(w, "# HELP packetrusher_procedure_started_total Procedure attempts started.\n# TYPE packetrusher_procedure_started_total counter")
	for _, r := range results {
		fmt.Fprintf(w, "packetrusher_procedure_started_total{procedure=%q} %d\n", r.Procedure, r.Started)
	}
	fmt.Fprintln(w, "# HELP packetrusher_procedure_completed_total Procedure attempts completed by outcome.\n# TYPE packetrusher_procedure_completed_total counter")
	for _, r := range results {
		fmt.Fprintf(w, "packetrusher_procedure_completed_total{procedure=%q,outcome=\"success\"} %d\n", r.Procedure, r.Success)
		fmt.Fprintf(w, "packetrusher_procedure_completed_total{procedure=%q,outcome=\"failure\"} %d\n", r.Procedure, r.Failure)
	}
	fmt.Fprintln(w, "# HELP packetrusher_procedure_pending Procedure attempts awaiting completion.\n# TYPE packetrusher_procedure_pending gauge")
	for _, r := range results {
		fmt.Fprintf(w, "packetrusher_procedure_pending{procedure=%q} %d\n", r.Procedure, r.Pending)
	}
	fmt.Fprintln(w, "# HELP packetrusher_procedure_duration_seconds Duration of successful procedure attempts.\n# TYPE packetrusher_procedure_duration_seconds histogram")
	for _, r := range results {
		for i, bound := range latencyBounds {
			fmt.Fprintf(w, "packetrusher_procedure_duration_seconds_bucket{procedure=%q,le=\"%g\"} %d\n", r.Procedure, bound, r.buckets[i])
		}
		fmt.Fprintf(w, "packetrusher_procedure_duration_seconds_bucket{procedure=%q,le=\"+Inf\"} %d\n", r.Procedure, r.Success)
		fmt.Fprintf(w, "packetrusher_procedure_duration_seconds_sum{procedure=%q} %g\n", r.Procedure, r.LatencySum)
		fmt.Fprintf(w, "packetrusher_procedure_duration_seconds_count{procedure=%q} %d\n", r.Procedure, r.Success)
	}
}
