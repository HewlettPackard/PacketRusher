// SPDX-License-Identifier: Apache-2.0
package main

import (
	"encoding/json"
	"fmt"
	"my5G-RANTester/internal/scenario"
	"os"
	"time"

	"github.com/urfave/cli/v2"
)

func controlCommands() []*cli.Command {
	return []*cli.Command{
		{Name: "control", Usage: "Inspect or trigger procedures in a running multi-ue test", Flags: []cli.Flag{
			&cli.PathFlag{Name: "socket", Required: true, Usage: "Local control socket created by --control-socket"},
			&cli.StringFlag{Name: "action", Value: "inspect", Usage: "inspect, wait, register, deregister, idle, reconnect, xn-handover or ng-handover"},
			&cli.IntFlag{Name: "ue", Usage: "PacketRusher UE ID (omit only for inspect)"},
			&cli.StringFlag{Name: "target", Usage: "Target configured gNB ID for handover"},
			&cli.DurationFlag{Name: "timeout", Value: 30 * time.Second, Usage: "Readiness and completion deadline (up to 2m)"},
		}, Action: func(c *cli.Context) error {
			request, err := controlRequest(c)
			if err != nil {
				return err
			}
			response, err := scenario.Call(c.Context, c.Path("socket"), request)
			if err != nil {
				return err
			}
			encoder := json.NewEncoder(c.App.Writer)
			encoder.SetIndent("", "  ")
			return encoder.Encode(response)
		}},
		{Name: "run-scenario", Usage: "Execute JSON procedure steps against a running test", Flags: []cli.Flag{
			&cli.PathFlag{Name: "socket", Required: true}, &cli.PathFlag{Name: "scenario", Required: true},
		}, Action: func(c *cli.Context) error {
			file, err := os.Open(c.Path("scenario"))
			if err != nil {
				return err
			}
			defer file.Close()
			plan, err := scenario.ReadPlan(file)
			if err != nil {
				return err
			}
			return plan.Run(c.Context, c.Path("socket"), c.App.Writer)
		}},
	}
}

func controlRequest(c *cli.Context) (scenario.Request, error) {
	duration := c.Duration("timeout")
	if duration < time.Millisecond || duration > 2*time.Minute {
		return scenario.Request{}, fmt.Errorf("--timeout must be between 1ms and 2m")
	}
	r := scenario.Request{UE: c.Int("ue"), Action: c.String("action"), Target: c.String("target"), TimeoutMS: int(duration / time.Millisecond)}
	return r, r.Validate()
}

func validateControl(c *cli.Context) error {
	if c.Args().Len() != 0 {
		return fmt.Errorf("unexpected positional arguments %q", c.Args().Slice())
	}
	if c.Path("socket") == "" {
		return fmt.Errorf("--socket is required")
	}
	if c.Command.Name == "control" {
		_, err := controlRequest(c)
		return err
	}
	return nil
}
