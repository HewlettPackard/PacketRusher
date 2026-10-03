// SPDX-License-Identifier: Apache-2.0
// Package testkit owns PacketRusher actors and nodes around a native aio5gc.
// Protocol handlers remain in aio5gc/msg; gates and lifecycle policy live here.
package testkit

import (
	"context"
	"fmt"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/tools"
	gnb "my5G-RANTester/internal/control_test_engine/gnb/context"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"my5G-RANTester/test/aio5gc"
	core "my5G-RANTester/test/aio5gc/context"
	coreTools "my5G-RANTester/test/aio5gc/lib/tools"
	"net/netip"
	"sync"
	"time"

	"github.com/free5gc/openapi/models"
)

// LocalConfig uses kernel-assigned SCTP AMF and gNB control ports. It keeps
// tunnel mode disabled: this fixture has no forwarding UPF or dataplane claim.
func LocalConfig() config.Config {
	return coreTools.GenerateDefaultConf(netip.MustParseAddrPort("127.0.0.1:0"), netip.MustParseAddrPort("127.0.0.1:2152"), []*config.AMF{{IPv4Port: config.IPv4Port{AddrPort: netip.MustParseAddrPort("127.0.0.1:0")}}})
}

type Fixture struct {
	Core        *core.Aio5gc
	Config      config.Config
	Gnbs        map[string]*gnb.GNBContext
	mu          sync.Mutex
	wg          sync.WaitGroup
	simulations []*tools.UESimulation
	closing     bool
	closed      chan struct{}
	closeErr    error
}

// Start binds the core, starts actual production gNBs and waits for observed
// NGSetup success. Callers can supply a builder with scenario-specific hooks.
func Start(ctx context.Context, builder *aio5gc.FiveGCBuilder, gnbCount int) (*Fixture, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if gnbCount < 1 {
		return nil, fmt.Errorf("at least one gNB is required")
	}
	core, err := builder.Build()
	if err != nil {
		return nil, err
	}
	f := &Fixture{Core: core, Config: core.Config(), closed: make(chan struct{})}
	f.Gnbs, err = tools.CreateGnbsContext(ctx, gnbCount, f.Config, &f.wg)
	if err == nil {
		err = tools.WaitGnbs(ctx, f.Gnbs)
	}
	if err != nil {
		cleanup, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = f.CloseContext(cleanup)
		return nil, err
	}
	return f, nil
}

// Provision creates a distinct subscriber from the fixture's UE configuration.
// It surfaces duplicate/invalid provisioning errors rather than ignoring them.
func (f *Fixture) Provision(id int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closing {
		return fmt.Errorf("fixture is stopping")
	}
	if id < 1 {
		return fmt.Errorf("UE ID must be positive")
	}
	cfg := f.Config.Ue
	security := core.SecurityContext{}
	security.SetMsin(tools.IncrementMsin(id, cfg.Msin))
	security.SetAuthSubscription(cfg.Key, cfg.Opc, "c9e8763286b5b9ffbdf56e1297d0887b", cfg.Amf, cfg.Sqn)
	security.SetAbba([]byte{0, 0})
	return f.Core.GetAMFContext().Provision(models.Snssai{Sst: int32(cfg.Snssai.Sst), Sd: cfg.Snssai.Sd}, security)
}

// StartUE uses the real simulation loop. Identity/timers/loop policy come from
// options; the owned core config and gNB registry cannot be accidentally mixed.
func (f *Fixture) StartUE(options tools.UESimulationConfig) (*tools.UESimulation, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closing {
		return nil, fmt.Errorf("fixture is stopping")
	}
	if options.UeId < 1 {
		return nil, fmt.Errorf("UE ID must be positive")
	}
	options.Cfg = core.CloneConfig(f.Config)
	options.Gnbs = f.Gnbs
	simulation := tools.SimulateSingleUE(options, &f.wg)
	f.simulations = append(f.simulations, simulation)
	return simulation, nil
}
func (f *Fixture) CloseContext(ctx context.Context) error {
	f.mu.Lock()
	if !f.closing {
		f.closing = true
		simulations := append([]*tools.UESimulation(nil), f.simulations...)
		// Cancellation first releases policy gates. Stopping the production gNBs
		// then cancels admitted and late UE connections before Kill/worker joining.
		f.Core.Stop()
		go func() {
			for _, node := range f.Gnbs {
				node.Terminate()
			}
			for _, simulation := range simulations {
				simulation.Send(procedures.UeTesterMessage{Type: procedures.Kill})
			}
			f.wg.Wait()
			err := f.Core.CloseContext(context.Background())
			f.mu.Lock()
			f.closeErr = err
			f.mu.Unlock()
			close(f.closed)
		}()
	}
	f.mu.Unlock()
	select {
	case <-f.closed:
		f.mu.Lock()
		defer f.mu.Unlock()
		return f.closeErr
	case <-ctx.Done():
		return ctx.Err()
	}
}
func (f *Fixture) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return f.CloseContext(ctx)
}
