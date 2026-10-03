// SPDX-License-Identifier: Apache-2.0
package context

import (
	"sync"
	"testing"
)

func sessionUE() *GNBUe {
	ue := &GNBUe{}
	ue.CreateUeContext("02f839", "imei", []string{"01"}, []string{"000001"}, nil)
	return ue
}

func createSession(ue *GNBUe, id int64) (*GnbPDUSession, error) {
	return ue.CreatePduSession(id, "10.0.0.1", "01", "000001", 0, 9, 1, 9, 100, 200)
}

func TestPduSessionMembershipSnapshotAndHandoverCopy(t *testing.T) {
	source, target := sessionUE(), sessionUE()
	// Existing callers may hold UE.Lock; accessors must use an independent lock.
	source.Lock()
	defer source.Unlock()
	session, err := createSession(source, 1)
	if err != nil {
		t.Fatal(err)
	}
	snapshot := source.GetPduSessions()
	target.CopyFromPreviousContext(source)
	if err := source.DeletePduSession(1); err != nil {
		t.Fatal(err)
	}
	if got, _ := source.GetPduSession(1); got != nil {
		t.Fatal("source session was not removed")
	}
	if got, _ := target.GetPduSession(1); got != session || snapshot[0] != session {
		t.Fatal("source deletion changed the copied membership or session identity")
	}
	if err := target.DeletePduSession(1); err != nil {
		t.Fatal(err)
	}
	if snapshot[0] != session {
		t.Fatal("target deletion changed an already returned snapshot")
	}
	source.SetPduSessions(snapshot)
	if sst, sd := source.GetSelectedNssai(1); sst != "01" || sd != "000001" {
		t.Fatalf("restored session has unexpected slice %s/%s", sst, sd)
	}
	source.CopyFromPreviousContext(source) // Self-copy must not recursively lock.
	if got, _ := source.GetPduSession(1); got != session {
		t.Fatal("self-copy changed session identity")
	}
}

func TestConcurrentPduSessionCreateHasOneWinner(t *testing.T) {
	ue := sessionUE()
	start := make(chan struct{})
	winners := make(chan *GnbPDUSession, 32)
	var workers sync.WaitGroup
	for i := 0; i < cap(winners); i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			if session, err := createSession(ue, 1); err == nil {
				winners <- session
			}
		}()
	}
	close(start)
	workers.Wait()
	close(winners)
	var winner *GnbPDUSession
	count := 0
	for session := range winners {
		winner = session
		count++
	}
	if count != 1 {
		t.Fatalf("duplicate creation had %d winners, want one", count)
	}
	if got, _ := ue.GetPduSession(1); got != winner {
		t.Fatal("stored session differs from successful creation")
	}
}

func TestConcurrentPduSessionAccessAndContextCopy(t *testing.T) {
	source, target := sessionUE(), sessionUE()
	shared, err := createSession(source, 15)
	if err != nil {
		t.Fatal(err)
	}
	target.CopyFromPreviousContext(source)
	start := make(chan struct{})
	var workers sync.WaitGroup
	run := func(work func()) {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			for i := 0; i < 500; i++ {
				work()
			}
		}()
	}
	run(func() {
		_, _ = createSession(source, 1)
		_ = source.DeletePduSession(1)
	})
	run(func() {
		source.CreateUeContext("020801", "imei", []string{"01"}, []string{"000001"}, nil)
	})
	run(func() {
		target.CopyFromPreviousContext(source)
		target.SetPduSessions(source.GetPduSessions())
	})
	run(func() {
		source.CopyFromPreviousContext(target) // Opposite-direction copy has no lock cycle.
	})
	run(func() {
		for _, ue := range []*GNBUe{source, target} {
			_, _ = ue.GetPduSession(1)
			_ = ue.GetPduSessions()
			_, _ = ue.GetSelectedNssai(1)
			_, _ = ue.GetUeMobility()
			_ = ue.GetUeMaskedImeiSv()
			_ = ue.GetUESecurityCapabilities()
		}
		_ = shared.GetUpfIp()
		_ = shared.GetTeidUplink()
		_ = shared.GetTeidDownlink()
	})
	run(func() {
		shared.SetUpfIp("10.0.0.2")
		shared.SetTeidUplink(101)
		shared.SetTeidDownlink(201)
	})
	close(start)
	workers.Wait()
	if snapshot := source.GetPduSessions(); snapshot[14] != shared {
		t.Fatal("stable session identity lost during concurrent context copies")
	}
}
