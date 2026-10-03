// SPDX-License-Identifier: Apache-2.0
package scenario

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"my5G-RANTester/internal/common/tools"
	gnb "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"net"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"
)

type Request struct {
	UE        int    `json:"ue"`
	Action    string `json:"action"`
	Target    string `json:"target,omitempty"`
	TimeoutMS int    `json:"timeout_ms,omitempty"`
}

func (r Request) Validate() error {
	switch r.Action {
	case "inspect":
		if r.UE < 0 {
			return errors.New("UE cannot be negative")
		}
	case "wait", "register", "deregister", "idle", "reconnect", "xn-handover", "ng-handover":
		if r.UE < 1 {
			return errors.New("an action requires a positive UE ID")
		}
	default:
		return fmt.Errorf("unknown control action %q", r.Action)
	}
	if r.TimeoutMS < 0 || r.TimeoutMS > 120000 {
		return errors.New("timeout_ms must be between 0 and 120000")
	}
	if r.Action == "xn-handover" || r.Action == "ng-handover" {
		if r.Target == "" {
			return errors.New("handover requires a target gNB ID")
		}
	} else if r.Target != "" {
		return errors.New("target applies only to handover")
	}
	return nil
}

func (r Request) Timeout() time.Duration {
	if r.TimeoutMS == 0 {
		return 30 * time.Second
	}
	return time.Duration(r.TimeoutMS) * time.Millisecond
}

type GNB struct {
	ID    string `json:"id"`
	Ready bool   `json:"ready"`
}
type Response struct {
	UEs   []procedures.Attachment `json:"ues"`
	GNBs  []GNB                   `json:"gnbs,omitempty"`
	Error string                  `json:"error,omitempty"`
}

type Registry struct {
	mu   sync.RWMutex
	ues  map[int]*tools.UESimulation
	gnbs map[string]*gnb.GNBContext
}

func NewRegistry(gnbs map[string]*gnb.GNBContext) *Registry {
	return &Registry{ues: make(map[int]*tools.UESimulation), gnbs: gnbs}
}
func (r *Registry) Add(id int, ue *tools.UESimulation) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ues[id] = ue
}

func (r *Registry) Execute(ctx context.Context, q Request) (Response, error) {
	response := Response{UEs: []procedures.Attachment{}}
	if err := q.Validate(); err != nil {
		return response, err
	}
	r.mu.RLock()
	selected := make(map[int]*tools.UESimulation)
	for id, ue := range r.ues {
		if q.UE == 0 || q.UE == id {
			selected[id] = ue
		}
	}
	r.mu.RUnlock()
	if q.UE != 0 && len(selected) == 0 {
		return response, fmt.Errorf("unknown UE %d", q.UE)
	}
	ids := make([]int, 0, len(selected))
	for id := range selected {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	for _, id := range ids {
		var a procedures.Attachment
		var err error
		if q.Action == "inspect" {
			a, err = selected[id].Inspect(ctx)
		} else {
			a, err = selected[id].Execute(ctx, q.Action, q.Target)
		}
		if errors.Is(err, procedures.ErrStopped) && q.Action == "inspect" {
			a = procedures.Attachment{UE: id, State: "stopped", ActivePDUSessions: []uint8{}}
			err = nil
		}
		response.UEs = append(response.UEs, a)
		if err != nil {
			return response, fmt.Errorf("UE %d %s: %w", id, q.Action, err)
		}
	}
	for id, node := range r.gnbs {
		response.GNBs = append(response.GNBs, GNB{id, node.NGSetupReady()})
	}
	sort.Slice(response.GNBs, func(i, j int) bool { return response.GNBs[i].ID < response.GNBs[j].ID })
	return response, nil
}

func Decode(reader io.Reader, value any) error {
	decoder := json.NewDecoder(io.LimitReader(reader, 1024*1024+1))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return errors.New("expected exactly one JSON value")
	}
	return nil
}

func (r *Registry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if req.URL.Path != "/control" {
		http.NotFound(w, req)
		return
	}
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var q Request
	if err := Decode(http.MaxBytesReader(w, req.Body, 1024*1024), &q); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(Response{Error: err.Error()})
		return
	}
	if err := q.Validate(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(Response{Error: err.Error()})
		return
	}
	ctx, cancel := context.WithTimeout(req.Context(), q.Timeout())
	defer cancel()
	response, err := r.Execute(ctx, q)
	if err != nil {
		response.Error = err.Error()
		w.WriteHeader(http.StatusConflict)
	}
	_ = json.NewEncoder(w).Encode(response)
}

type Server struct {
	server    *http.Server
	listener  *net.UnixListener
	cancel    context.CancelFunc
	path      string
	identity  os.FileInfo
	closeOnce sync.Once
	closeErr  error
}

func Listen(path string, registry *Registry) (*Server, error) {
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: path, Net: "unix"})
	if err != nil {
		return nil, fmt.Errorf("listen for scenario control: %w", err)
	}
	listener.SetUnlinkOnClose(false)
	identity, statErr := os.Lstat(path)
	if statErr != nil {
		listener.Close()
		return nil, statErr
	}
	cleanup := func() {
		listener.Close()
		if current, e := os.Lstat(path); e == nil && os.SameFile(identity, current) {
			_ = os.Remove(path)
		}
	}
	if err = os.Chmod(path, 0600); err != nil {
		cleanup()
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	s := &Server{listener: listener, cancel: cancel, path: path, identity: identity}
	s.server = &http.Server{Handler: registry, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 130 * time.Second, IdleTimeout: 30 * time.Second, BaseContext: func(net.Listener) context.Context { return ctx }}
	go func() { _ = s.server.Serve(listener) }()
	return s, nil
}

func (s *Server) Close() error {
	s.closeOnce.Do(func() {
		s.cancel()
		s.closeErr = s.server.Close()
		if current, e := os.Lstat(s.path); e == nil && current.Mode()&os.ModeSocket != 0 && os.SameFile(s.identity, current) {
			s.closeErr = errors.Join(s.closeErr, os.Remove(s.path))
		}
	})
	return s.closeErr
}

func Call(ctx context.Context, socket string, q Request) (Response, error) {
	if err := q.Validate(); err != nil {
		return Response{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, q.Timeout()+time.Second)
	defer cancel()
	transport := &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "unix", socket)
	}}
	defer transport.CloseIdleConnections()
	body, _ := json.Marshal(q)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost/control", bytes.NewReader(body))
	if err != nil {
		return Response{}, err
	}
	resp, err := (&http.Client{Transport: transport}).Do(req)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()
	var result Response
	if err := Decode(resp.Body, &result); err != nil {
		return result, err
	}
	if result.Error != "" {
		return result, errors.New(result.Error)
	}
	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("control returned HTTP %d", resp.StatusCode)
	}
	return result, nil
}
