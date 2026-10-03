// SPDX-License-Identifier: Apache-2.0
package context

import (
	stdcontext "context"
	"fmt"
	"sort"
	"sync"

	"github.com/free5gc/util/fsm"
)

// Snapshot contains values, never live UE/SM pointers. States are last observed
// FSM entries, not a promise that the client has accepted the corresponding
// wire response. Entries preserve transition counts even after PDU retirement.
type AssociationRecord struct {
	Local  string
	Remote string
}

type RetiredAssociationRecord struct {
	AssociationRecord
	State uint32
	Cause string
}

type Snapshot struct {
	Associations        []AssociationRecord
	UEs                 []UERecord
	Errors              []error
	RetiredAssociations []RetiredAssociationRecord
}
type UERecord struct {
	AMFID    int64
	MSIN     string
	State    fsm.StateType
	Entries  map[fsm.StateType]uint64
	Sessions map[int32]PDURecord
}
type PDURecord struct {
	State   fsm.StateType
	Entries map[fsm.StateType]uint64
}
type observations struct {
	mu      sync.Mutex
	changed chan struct{}
	ues     map[int64]UERecord
	errors  []error
	retired []RetiredAssociationRecord
}

func (o *observations) initLocked() {
	if o.changed == nil {
		o.changed = make(chan struct{})
		o.ues = make(map[int64]UERecord)
	}
}
func (o *observations) signalLocked() { close(o.changed); o.changed = make(chan struct{}) }
func copyEntries(m map[fsm.StateType]uint64) map[fsm.StateType]uint64 {
	n := make(map[fsm.StateType]uint64, len(m))
	for k, v := range m {
		n[k] = v
	}
	return n
}
func (o *observations) snapshotLocked() Snapshot {
	s := Snapshot{Errors: append([]error(nil), o.errors...), RetiredAssociations: append([]RetiredAssociationRecord(nil), o.retired...)}
	for _, ue := range o.ues {
		ue.Entries = copyEntries(ue.Entries)
		sessions := make(map[int32]PDURecord, len(ue.Sessions))
		for id, pdu := range ue.Sessions {
			pdu.Entries = copyEntries(pdu.Entries)
			sessions[id] = pdu
		}
		ue.Sessions = sessions
		s.UEs = append(s.UEs, ue)
	}
	sort.Slice(s.UEs, func(i, j int) bool { return s.UEs[i].AMFID < s.UEs[j].AMFID })
	return s
}
func (a *Aio5gc) Snapshot() Snapshot {
	o := &a.observations
	o.mu.Lock()
	defer o.mu.Unlock()
	o.initLocked()
	snapshot := o.snapshotLocked()
	snapshot.Associations = a.amfContext.Associations()
	return snapshot
}

// Wait observes a snapshot and its next-change signal atomically, avoiding both
// polling sleeps and lost wakeups. Predicate runs outside the observation lock.
func (a *Aio5gc) Wait(ctx stdcontext.Context, predicate func(Snapshot) bool) (Snapshot, error) {
	for {
		o := &a.observations
		o.mu.Lock()
		o.initLocked()
		snapshot, changed := o.snapshotLocked(), o.changed
		snapshot.Associations = a.amfContext.Associations()
		o.mu.Unlock()
		if predicate(snapshot) {
			return snapshot, nil
		}
		select {
		case <-changed:
		case <-ctx.Done():
			return snapshot, fmt.Errorf("wait for core state: %w; last snapshot: %+v", ctx.Err(), snapshot)
		case <-a.Context().Done():
			return snapshot, stdcontext.Canceled
		}
	}
}

// RecordError makes protocol/hook errors visible to assertions, not only logs.
func (a *Aio5gc) RecordError(err error) {
	if err == nil {
		return
	}
	o := &a.observations
	o.mu.Lock()
	defer o.mu.Unlock()
	o.initLocked()
	o.errors = append(o.errors, err)
	o.signalLocked()
}

func (a *Aio5gc) RecordRetiredAssociation(gnb *GNBContext, retired *RetiredNGSetup) {
	local, remote := gnb.Endpoints()
	o := &a.observations
	o.mu.Lock()
	defer o.mu.Unlock()
	o.initLocked()
	o.retired = append(o.retired, RetiredAssociationRecord{AssociationRecord: AssociationRecord{Local: local, Remote: remote}, State: retired.State, Cause: fmt.Sprint(retired.Cause)})
	o.signalLocked()
}
func (a *Aio5gc) observeCreatedUE(ue *UEContext) {
	o := &a.observations
	o.mu.Lock()
	defer o.mu.Unlock()
	o.initLocked()
	o.ues[ue.amfNgapId] = UERecord{AMFID: ue.amfNgapId, State: Deregistered, Entries: make(map[fsm.StateType]uint64), Sessions: make(map[int32]PDURecord)}
	o.signalLocked()
}
func (a *Aio5gc) observedCallbacks(callbacks fsm.Callbacks, pdu bool) fsm.Callbacks {
	out := make(fsm.Callbacks, len(callbacks))
	for state, cb := range callbacks {
		out[state] = cb
	}
	states := []fsm.StateType{AuthenticationInitiated, Authenticated, Registered, DeregisteredInitiated, Deregistered}
	if pdu {
		states = []fsm.StateType{Inactive, InactivePending, Active, ModificationPending}
	}
	for _, state := range states {
		cb := out[state]
		out[state] = func(st *fsm.State, event fsm.EventType, args fsm.ArgsType) {
			if event == fsm.EntryEvent {
				if ue, ok := args["ue"].(*UEContext); ok {
					o := &a.observations
					o.mu.Lock()
					o.initLocked()
					record := o.ues[ue.amfNgapId]
					if record.Entries == nil {
						record = UERecord{AMFID: ue.amfNgapId, Entries: make(map[fsm.StateType]uint64), Sessions: make(map[int32]PDURecord)}
					}
					if sc := ue.GetSecurityContext(); sc != nil {
						record.MSIN = sc.GetMsin()
					}
					if pdu {
						if sm, ok := args["sm"].(*SmContext); ok {
							session := record.Sessions[sm.GetPduSessionId()]
							if session.Entries == nil {
								session.Entries = make(map[fsm.StateType]uint64)
							}
							session.State = state
							session.Entries[state]++
							record.Sessions[sm.GetPduSessionId()] = session
						}
					} else {
						record.State = state
						record.Entries[state]++
					}
					o.ues[ue.amfNgapId] = record
					o.signalLocked()
					o.mu.Unlock()
				}
			}
			if cb != nil {
				cb(st, event, args)
			}
		}
	}
	return out
}

func (a *Aio5gc) observeAssociationChange() {
	o := &a.observations
	o.mu.Lock()
	defer o.mu.Unlock()
	o.initLocked()
	o.signalLocked()
}
