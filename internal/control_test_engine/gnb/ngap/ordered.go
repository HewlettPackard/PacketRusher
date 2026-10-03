// SPDX-License-Identifier: Apache-2.0

package ngap

import (
	"context"
	"sync"

	"github.com/free5gc/ngap/ie"
	ngapmsg "github.com/free5gc/ngap/message"
	gnbcontext "my5G-RANTester/internal/control_test_engine/gnb/context"
)

// orderedDispatcher runs each UE's messages in receive order. Other UEs progress
// independently. The global limit provides backpressure without dropping PDUs;
// idle workers are removed rather than retained for every UE ever registered.
type orderedDispatcher struct {
	mu        sync.Mutex
	queues    map[int64][]func()
	ueAliases map[int64]int64
	aliasUEs  map[int64]int64
	slots     chan struct{}
	workers   sync.WaitGroup
	stopped   bool
	done      chan struct{}
}

func newOrderedDispatcher(limit int) *orderedDispatcher {
	return &orderedDispatcher{queues: make(map[int64][]func()), ueAliases: make(map[int64]int64), aliasUEs: make(map[int64]int64), slots: make(chan struct{}, limit), done: make(chan struct{})}
}

func (d *orderedDispatcher) enqueue(key int64, work func()) bool {
	return d.enqueueContext(context.Background(), key, work)
}

func (d *orderedDispatcher) enqueueContext(ctx context.Context, key int64, work func()) bool {
	select {
	case d.slots <- struct{}{}:
	case <-ctx.Done():
		return false
	case <-d.done:
		return false
	}
	d.mu.Lock()
	if d.stopped {
		d.mu.Unlock()
		<-d.slots
		return false
	}
	_, running := d.queues[key]
	d.queues[key] = append(d.queues[key], work)
	if !running {
		d.workers.Add(1)
		go d.run(key)
	}
	d.mu.Unlock()
	return true
}

func (d *orderedDispatcher) stop() {
	d.mu.Lock()
	if !d.stopped {
		d.stopped = true
		close(d.done)
	}
	d.mu.Unlock()
}

func (d *orderedDispatcher) run(key int64) {
	defer d.workers.Done()
	for {
		d.mu.Lock()
		queue := d.queues[key]
		if len(queue) == 0 {
			delete(d.queues, key)
			if ran, ok := d.aliasUEs[key]; ok {
				if d.ueAliases[ran] == key {
					delete(d.ueAliases, ran)
				}
				delete(d.aliasUEs, key)
			}
			d.mu.Unlock()
			return
		}
		work := queue[0]
		queue[0] = nil
		d.queues[key] = queue[1:]
		d.mu.Unlock()
		work()
		<-d.slots
	}
}

// Called after the association reader stops enqueueing, before releasing its UEs.
func (d *orderedDispatcher) wait() { d.workers.Wait() }

func messageUEIDs(message ngapmsg.Message) (ran, amf int64, hasAMF bool) {
	var ranID *ie.RANUENGAPID
	var amfID *ie.AMFUENGAPID
	switch value := message.(type) {
	case *ngapmsg.DownlinkNASTransport:
		ranID, amfID = value.RANUENGAPID, value.AMFUENGAPID
	case *ngapmsg.InitialContextSetupRequest:
		ranID, amfID = value.RANUENGAPID, value.AMFUENGAPID
	case *ngapmsg.PDUSessionResourceSetupRequest:
		ranID, amfID = value.RANUENGAPID, value.AMFUENGAPID
	case *ngapmsg.PDUSessionResourceReleaseCommand:
		ranID, amfID = value.RANUENGAPID, value.AMFUENGAPID
	case *ngapmsg.ErrorIndication:
		ranID, amfID = value.RANUENGAPID, value.AMFUENGAPID
	case *ngapmsg.PathSwitchRequestAcknowledge:
		ranID, amfID = value.RANUENGAPID, value.AMFUENGAPID
	case *ngapmsg.HandoverCommand:
		ranID, amfID = value.RANUENGAPID, value.AMFUENGAPID
	case *ngapmsg.HandoverRequest:
		amfID = value.AMFUENGAPID
	case *ngapmsg.UEContextReleaseCommand:
		if value.UENGAPIDs != nil {
			switch ids := value.UENGAPIDs.Choice.(type) {
			case *ie.UENGAPIDPair:
				ranID, amfID = ids.RANUENGAPID, ids.AMFUENGAPID
			case *ie.AMFUENGAPID:
				amfID = ids
			}
		}
	}
	if ranID != nil {
		ran = ranID.Value
	}
	if amfID != nil {
		amf, hasAMF = amfID.Value, true
	}
	return
}

// The AMF ID also appears in AMF-only release commands. Route at receipt time,
// independently of whether the previous handler has published it into context.
func messageUEKey(message ngapmsg.Message) int64 {
	ran, amf, hasAMF := messageUEIDs(message)
	if hasAMF {
		return -amf - 1
	}
	return ran
}

// A RAN-only procedure must follow earlier PDUs carrying both IDs, even before
// the first handler publishes the AMF ID. Temporary aliases live only while the
// worker is active; once it idles, context publication supplies the identity.
func (d *orderedDispatcher) messageKey(gnb *gnbcontext.GNBContext, message ngapmsg.Message) int64 {
	ran, amf, hasAMF := messageUEIDs(message)
	key := ran
	d.mu.Lock()
	defer d.mu.Unlock()
	if hasAMF {
		key = -amf - 1
	} else if ran != 0 {
		if alias, ok := d.ueAliases[ran]; ok {
			key = alias
		} else if ue, err := gnb.GetGnbUe(ran); err == nil && ue.HasAmfUeId() {
			key = -ue.GetAmfUeId() - 1
		}
	}
	if ran != 0 && key < 0 {
		d.ueAliases[ran] = key
		d.aliasUEs[key] = ran
	}
	return key
}
