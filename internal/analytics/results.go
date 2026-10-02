// SPDX-License-Identifier: Apache-2.0
package analytics

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type Procedure string
type Outcome string

const (
	Registration         Procedure = "registration"
	SessionEstablishment Procedure = "pdu_session_establishment"
	Success              Outcome   = "success"
	Failure              Outcome   = "failure"
	Cancelled            Outcome   = "cancelled"
)

// Latency buckets stay bounded regardless of the number of UEs or loop iterations.
var latencyBounds = [...]float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}

type attemptKey struct {
	ue        int64
	session   uint8
	procedure Procedure
}
type aggregate struct {
	Started      uint64                     `json:"started"`
	Success      uint64                     `json:"success"`
	Failure      uint64                     `json:"failure"`
	Cancelled    uint64                     `json:"cancelled"`
	Pending      uint64                     `json:"pending"`
	LatencyCount uint64                     `json:"latency_count"`
	LatencySum   float64                    `json:"latency_seconds_sum"`
	LatencyMin   float64                    `json:"latency_seconds_min"`
	LatencyMax   float64                    `json:"latency_seconds_max"`
	LatencyMean  float64                    `json:"latency_seconds_mean"`
	Buckets      [len(latencyBounds)]uint64 `json:"-"`
}

type ProcedureResult struct {
	Procedure Procedure `json:"procedure"`
	aggregate
}

type Report struct {
	SchemaVersion int               `json:"schema_version"`
	StartedAt     time.Time         `json:"started_at"`
	EndedAt       *time.Time        `json:"ended_at,omitempty"`
	Procedures    []ProcedureResult `json:"procedures"`
}

// Recorder counts procedure attempts. No subscriber identity becomes a metric
// label or report field. The only per-UE storage is for currently pending work.
type Recorder struct {
	mu      sync.Mutex
	now     func() time.Time
	started time.Time
	ended   *time.Time
	pending map[attemptKey]time.Time
	totals  map[Procedure]*aggregate
}

var current atomic.Pointer[Recorder]

func Current() *Recorder     { return current.Load() }
func SetCurrent(r *Recorder) { current.Store(r) }

func NewRecorder() *Recorder { return newRecorder(time.Now) }
func newRecorder(now func() time.Time) *Recorder {
	return &Recorder{now: now, started: now().UTC(), pending: make(map[attemptKey]time.Time), totals: map[Procedure]*aggregate{
		Registration: {}, SessionEstablishment: {},
	}}
}

func (r *Recorder) Begin(ue int64, session uint8, procedure Procedure) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.ended != nil {
		return
	}
	k := attemptKey{ue, session, procedure}
	if _, exists := r.pending[k]; exists {
		return
	}
	a := r.totals[procedure]
	if a == nil {
		return
	}
	r.pending[k] = r.now()
	a.Started++
	a.Pending++
}

func (r *Recorder) Finish(ue int64, session uint8, procedure Procedure, outcome Outcome) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.finish(attemptKey{ue, session, procedure}, outcome)
}

func (r *Recorder) finish(k attemptKey, outcome Outcome) {
	start, exists := r.pending[k]
	if !exists || (outcome != Success && outcome != Failure && outcome != Cancelled) {
		return
	}
	delete(r.pending, k)
	a := r.totals[k.procedure]
	a.Pending--
	switch outcome {
	case Success:
		a.Success++
		seconds := r.now().Sub(start).Seconds()
		if seconds < 0 {
			seconds = 0
		}
		if a.LatencyCount == 0 || seconds < a.LatencyMin {
			a.LatencyMin = seconds
		}
		if seconds > a.LatencyMax {
			a.LatencyMax = seconds
		}
		a.LatencyCount++
		a.LatencySum += seconds
		a.LatencyMean = a.LatencySum / float64(a.LatencyCount)
		for i, bound := range latencyBounds {
			if seconds <= bound {
				a.Buckets[i]++
			}
		}
	case Failure:
		a.Failure++
	case Cancelled:
		a.Cancelled++
	}
}

func (r *Recorder) CancelUE(ue int64) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for k := range r.pending {
		if k.ue == ue {
			r.finish(k, Cancelled)
		}
	}
}

// Close freezes the report and accounts for all remaining work as cancellation.
func (r *Recorder) Close() {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.ended != nil {
		return
	}
	for k := range r.pending {
		r.finish(k, Cancelled)
	}
	end := r.now().UTC()
	r.ended = &end
}

func (r *Recorder) Snapshot() Report {
	r.mu.Lock()
	defer r.mu.Unlock()
	report := Report{SchemaVersion: 1, StartedAt: r.started, Procedures: make([]ProcedureResult, 0, len(r.totals))}
	if r.ended != nil {
		end := *r.ended
		report.EndedAt = &end
	}
	for procedure, a := range r.totals {
		report.Procedures = append(report.Procedures, ProcedureResult{procedure, *a})
	}
	sort.Slice(report.Procedures, func(i, j int) bool { return report.Procedures[i].Procedure < report.Procedures[j].Procedure })
	return report
}

func (r *Recorder) WriteJSON(w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(r.Snapshot())
}

func (r *Recorder) WriteCSV(w io.Writer) error {
	writer := csv.NewWriter(w)
	if err := writer.Write([]string{"procedure", "started", "success", "failure", "cancelled", "pending", "latency_count", "latency_seconds_mean", "latency_seconds_min", "latency_seconds_max"}); err != nil {
		return err
	}
	for _, a := range r.Snapshot().Procedures {
		row := []string{string(a.Procedure)}
		for _, n := range []uint64{a.Started, a.Success, a.Failure, a.Cancelled, a.Pending, a.LatencyCount} {
			row = append(row, strconv.FormatUint(n, 10))
		}
		for _, n := range []float64{a.LatencyMean, a.LatencyMin, a.LatencyMax} {
			row = append(row, strconv.FormatFloat(n, 'g', -1, 64))
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

// ServeHTTP exposes Prometheus text format without requiring a NAS codec version.
// Latency histograms contain successful procedures only, with bounded labels.
func (r *Recorder) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/metrics" {
		http.NotFound(w, request)
		return
	}
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	report := r.Snapshot()
	fmt.Fprintln(w, "# HELP packetrusher_procedure_started_total Procedure attempts started.")
	fmt.Fprintln(w, "# TYPE packetrusher_procedure_started_total counter")
	for _, a := range report.Procedures {
		fmt.Fprintf(w, "packetrusher_procedure_started_total{procedure=%q} %d\n", a.Procedure, a.Started)
	}
	fmt.Fprintln(w, "# HELP packetrusher_procedure_completed_total Procedure attempts completed by outcome.")
	fmt.Fprintln(w, "# TYPE packetrusher_procedure_completed_total counter")
	for _, a := range report.Procedures {
		for _, count := range []struct {
			outcome Outcome
			n       uint64
		}{{Success, a.Success}, {Failure, a.Failure}, {Cancelled, a.Cancelled}} {
			fmt.Fprintf(w, "packetrusher_procedure_completed_total{procedure=%q,outcome=%q} %d\n", a.Procedure, count.outcome, count.n)
		}
	}
	fmt.Fprintln(w, "# HELP packetrusher_procedure_pending Procedure attempts awaiting completion.")
	fmt.Fprintln(w, "# TYPE packetrusher_procedure_pending gauge")
	for _, a := range report.Procedures {
		fmt.Fprintf(w, "packetrusher_procedure_pending{procedure=%q} %d\n", a.Procedure, a.Pending)
	}
	fmt.Fprintln(w, "# HELP packetrusher_procedure_duration_seconds Duration of successful procedure attempts.")
	fmt.Fprintln(w, "# TYPE packetrusher_procedure_duration_seconds histogram")
	for _, a := range report.Procedures {
		for i, bound := range latencyBounds {
			fmt.Fprintf(w, "packetrusher_procedure_duration_seconds_bucket{procedure=%q,le=%q} %d\n", a.Procedure, strconv.FormatFloat(bound, 'g', -1, 64), a.Buckets[i])
		}
		fmt.Fprintf(w, "packetrusher_procedure_duration_seconds_bucket{procedure=%q,le=\"+Inf\"} %d\n", a.Procedure, a.LatencyCount)
		fmt.Fprintf(w, "packetrusher_procedure_duration_seconds_sum{procedure=%q} %g\n", a.Procedure, a.LatencySum)
		fmt.Fprintf(w, "packetrusher_procedure_duration_seconds_count{procedure=%q} %d\n", a.Procedure, a.LatencyCount)
	}
}
