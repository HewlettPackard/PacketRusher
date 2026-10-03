// SPDX-License-Identifier: Apache-2.0

package ngap

import (
	"context"
	"sync"

	"github.com/free5gc/ngap/ngapType"
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

func messageUEIDs(pdu *ngapType.NGAPPDU) (ran, amf int64, hasAMF bool) {
	switch pdu.Present {
	case ngapType.NGAPPDUPresentInitiatingMessage:
		if pdu.InitiatingMessage == nil {
			return 0, 0, false
		}
		value := pdu.InitiatingMessage.Value
		if value.DownlinkNASTransport != nil {
			for _, ie := range value.DownlinkNASTransport.ProtocolIEs.List {
				if ie.Value.RANUENGAPID != nil {
					ran = ie.Value.RANUENGAPID.Value
				}
				if ie.Value.AMFUENGAPID != nil {
					amf = ie.Value.AMFUENGAPID.Value
					hasAMF = true
				}
			}
		} else if value.InitialContextSetupRequest != nil {
			for _, ie := range value.InitialContextSetupRequest.ProtocolIEs.List {
				if ie.Value.RANUENGAPID != nil {
					ran = ie.Value.RANUENGAPID.Value
				}
				if ie.Value.AMFUENGAPID != nil {
					amf = ie.Value.AMFUENGAPID.Value
					hasAMF = true
				}
			}
		} else if value.PDUSessionResourceSetupRequest != nil {
			for _, ie := range value.PDUSessionResourceSetupRequest.ProtocolIEs.List {
				if ie.Value.RANUENGAPID != nil {
					ran = ie.Value.RANUENGAPID.Value
				}
				if ie.Value.AMFUENGAPID != nil {
					amf = ie.Value.AMFUENGAPID.Value
					hasAMF = true
				}
			}
		} else if value.PDUSessionResourceReleaseCommand != nil {
			for _, ie := range value.PDUSessionResourceReleaseCommand.ProtocolIEs.List {
				if ie.Value.RANUENGAPID != nil {
					ran = ie.Value.RANUENGAPID.Value
				}
				if ie.Value.AMFUENGAPID != nil {
					amf = ie.Value.AMFUENGAPID.Value
					hasAMF = true
				}
			}
		} else if value.ErrorIndication != nil {
			for _, ie := range value.ErrorIndication.ProtocolIEs.List {
				if ie.Value.RANUENGAPID != nil {
					ran = ie.Value.RANUENGAPID.Value
				}
				if ie.Value.AMFUENGAPID != nil {
					amf = ie.Value.AMFUENGAPID.Value
					hasAMF = true
				}
			}
		} else if value.UEContextReleaseCommand != nil {
			for _, ie := range value.UEContextReleaseCommand.ProtocolIEs.List {
				ids := ie.Value.UENGAPIDs
				if ids == nil {
					continue
				}
				if ids.UENGAPIDPair != nil {
					ran = ids.UENGAPIDPair.RANUENGAPID.Value
					amf = ids.UENGAPIDPair.AMFUENGAPID.Value
					hasAMF = true
				} else if ids.AMFUENGAPID != nil {
					amf = ids.AMFUENGAPID.Value
					hasAMF = true
				}
			}
		} else if value.HandoverRequest != nil {
			for _, ie := range value.HandoverRequest.ProtocolIEs.List {
				if ie.Value.AMFUENGAPID != nil {
					amf = ie.Value.AMFUENGAPID.Value
					hasAMF = true
				}
			}
		}
	case ngapType.NGAPPDUPresentSuccessfulOutcome:
		if pdu.SuccessfulOutcome == nil {
			return 0, 0, false
		}
		value := pdu.SuccessfulOutcome.Value
		if value.PathSwitchRequestAcknowledge != nil {
			for _, ie := range value.PathSwitchRequestAcknowledge.ProtocolIEs.List {
				if ie.Value.RANUENGAPID != nil {
					ran = ie.Value.RANUENGAPID.Value
				}
				if ie.Value.AMFUENGAPID != nil {
					amf = ie.Value.AMFUENGAPID.Value
					hasAMF = true
				}
			}
		} else if value.HandoverCommand != nil {
			for _, ie := range value.HandoverCommand.ProtocolIEs.List {
				if ie.Value.RANUENGAPID != nil {
					ran = ie.Value.RANUENGAPID.Value
				}
				if ie.Value.AMFUENGAPID != nil {
					amf = ie.Value.AMFUENGAPID.Value
					hasAMF = true
				}
			}
		}
	}
	return ran, amf, hasAMF
}

// UE-associated procedures carry the AMF ID, including the AMF-only release
// CHOICE. Prefer it so routing never depends on whether an earlier handler has
// already published that ID into the UE context.
func messageUEKey(pdu *ngapType.NGAPPDU) int64 {
	ran, amf, hasAMF := messageUEIDs(pdu)
	if hasAMF {
		return -amf - 1
	}
	return ran
}

// A RAN-only procedure must follow earlier PDUs carrying both IDs, even before
// the first handler publishes the AMF ID. Temporary aliases live only while the
// worker is active; once it idles, context publication supplies the identity.
func (d *orderedDispatcher) messageKey(gnb *gnbcontext.GNBContext, message *ngapType.NGAPPDU) int64 {
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
