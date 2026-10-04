/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */
package main

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/urfave/cli/v2"
)

// Replace only command actions so these tests exercise the real parser and
// validation hooks without loading a config or opening telecom sockets.
func parserApp(t *testing.T, action cli.ActionFunc) *cli.App {
	t.Helper()
	app := newApp()
	app.Writer = io.Discard
	app.ErrWriter = io.Discard
	for _, command := range app.Commands {
		command.Action = action
	}
	return app
}

func TestTunnelBooleanFlagsInEitherOrder(t *testing.T) {
	for _, flags := range [][]string{
		{"--tunnel", "-d", "--tunnel-vrf=false"},
		{"-d", "--tunnel", "--tunnel-vrf=false"},
		{"--tunnel=true", "--tunnel-vrf=false", "--dedicatedGnb"},
	} {
		t.Run(strings.Join(flags, " "), func(t *testing.T) {
			called := false
			app := parserApp(t, func(c *cli.Context) error {
				called = true
				if !c.Bool("tunnel") || !c.Bool("dedicatedGnb") || c.Bool("tunnel-vrf") {
					t.Fatalf("unexpected flags: tunnel=%v dedicated=%v vrf=%v", c.Bool("tunnel"), c.Bool("dedicatedGnb"), c.Bool("tunnel-vrf"))
				}
				return nil
			})
			args := append([]string{"packetrusher", "multi-ue", "-n", "2"}, flags...)
			if err := app.Run(args); err != nil {
				t.Fatal(err)
			}
			if !called {
				t.Fatal("command action was not called")
			}
		})
	}
}

func TestRejectArgumentsBeforeStartingCommand(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"separate tunnel boolean", []string{"multi-ue", "-n", "2", "--tunnel", "true", "-d", "--tunnel-vrf", "false"}, "a boolean flag takes its value as --flag=false"},
		{"separate false boolean", []string{"multi-ue", "-n", "2", "-d", "--tunnel-vrf", "false"}, "a boolean flag takes its value as --flag=false"},
		{"unexpected argument", []string{"gnb", "extra"}, `unexpected arguments ["extra"]`},
		{"zero UEs", []string{"multi-ue", "-n", "0"}, "--number-of-ues must be at least 1"},
		{"negative UEs", []string{"multi-ue", "-n=-1"}, "--number-of-ues must be at least 1"},
		{"a tunnel without PDU session", []string{"multi-ue", "-n", "1", "--tunnel", "--numPduSessions=0"}, "at least 1 with --tunnel"},
		{"too many PDU sessions", []string{"multi-ue", "-n", "1", "--numPduSessions=16"}, "--numPduSessions must be between 0 and 15"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			app := parserApp(t, func(*cli.Context) error { called = true; return nil })
			err := app.Run(append([]string{"packetrusher", "--config", "/nonexistent/config.yml"}, tc.args...))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error=%v, want containing %q", err, tc.want)
			}
			if called {
				t.Fatal("invalid arguments reached the command action")
			}
		})
	}
}

// "packetrusher <command> help" prints the help of the command, as --help does.
func TestCommandHelp(t *testing.T) {
	for _, command := range []string{"ue", "multi-ue"} {
		var output bytes.Buffer
		app := parserApp(t, func(*cli.Context) error { t.Fatal("help started the command"); return nil })
		app.Writer = &output
		if err := app.Run([]string{"packetrusher", command, "help"}); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(output.String(), "packetrusher "+command) || !strings.Contains(output.String(), "OPTIONS:") {
			t.Fatalf("%s help printed %q", command, output.String())
		}
	}
}
