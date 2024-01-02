/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package templates

import (
	"my5G-RANTester/config"
	"my5G-RANTester/internal/control_test_engine/gnb"
	"my5G-RANTester/internal/control_test_engine/procedures"
	"my5G-RANTester/internal/control_test_engine/ue"
	"my5G-RANTester/internal/script"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/tetratelabs/wazero"
)

func TestWithCustomScenario(scenarioPath string) {
	wg := sync.WaitGroup{}

	cfg := config.GetConfig()

	wg.Add(1)

	gnb := gnb.InitGnb(cfg, &wg)

	time.Sleep(1 * time.Second)

	ueChan := make(chan procedures.UeTesterMessage)

	wg.Add(1)

	_ = ue.NewUE(cfg, 1, ueChan, gnb.GetInboundChannel(), &wg)

	ctx, runtime := script.NewCustomScenario(scenarioPath)

	_, err := runtime.NewHostModuleBuilder("env").
		NewFunctionBuilder().
		WithFunc(func(ueId uint32) {
			ueChan <- procedures.UeTesterMessage{Type: procedures.Registration}
		}).
		Export("attach").
		NewFunctionBuilder().
		WithFunc(func(ueId uint32) {
			ueChan <- procedures.UeTesterMessage{Type: procedures.Deregistration}
		}).
		Export("detach").
		NewFunctionBuilder().
		WithFunc(func(ueId uint32, pduSessionId uint8) {
			ueChan <- procedures.UeTesterMessage{Type: procedures.NewPDUSession, Param: pduSessionId - 1}
		}).
		Export("pduSessionRequest").
		NewFunctionBuilder().
		WithFunc(func(ueId uint32, pduSessionId uint8) {
			ueChan <- procedures.UeTesterMessage{Type: procedures.DestroyPDUSession, Param: pduSessionId - 1}
		}).
		Export("pduSessionRelease").
		NewFunctionBuilder().
		WithFunc(func(v uint32) {
			time.Sleep(time.Duration(v) * time.Millisecond)
		}).
		Export("think").
		Instantiate(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Instantiate the guest Wasm into the same runtime. It exports the `add`
	// function, implemented in WebAssembly.
	addWasm, err := os.ReadFile(scenarioPath)
	if err != nil {
		log.Fatal(err)
	}

	config := wazero.NewModuleConfig().
		WithStdout(os.Stdout).
		WithStderr(os.Stderr)

	module, err := runtime.InstantiateWithConfig(ctx, addWasm, config)
	if err != nil {
		log.Fatal("failed to instantiate module: %v", err)
	}

	// Call the `add` function and print the results to the console.
	ueHandler := module.ExportedFunction("ueHandler")
	results, err := ueHandler.Call(ctx, 1)
	if err != nil {
		log.Fatal("failed to call ", err)
	}

	log.Info(results)

	wg.Wait()
}
