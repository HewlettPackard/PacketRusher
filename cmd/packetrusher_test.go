package main

import (
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
		{"separate tunnel boolean", []string{"multi-ue", "-n", "2", "--tunnel", "true", "-d", "--tunnel-vrf", "false"}, "boolean flags take no separate value"},
		{"separate false boolean", []string{"multi-ue", "-n", "2", "-d", "--tunnel-vrf", "false"}, "boolean flags take no separate value"},
		{"unexpected argument", []string{"gnb", "extra"}, "unexpected positional arguments"},
		{"zero UEs", []string{"multi-ue", "-n", "0"}, "--number-of-ues must be at least 1"},
		{"negative UEs", []string{"multi-ue", "-n=-1"}, "--number-of-ues must be at least 1"},
		{"negative PDU sessions", []string{"multi-ue", "-n", "1", "--numPduSessions=-1"}, "--numPduSessions must be between 0 and 15"},
		{"zero sessions with tunnel", []string{"multi-ue", "-n", "1", "--numPduSessions=0", "--tunnel"}, "--tunnel requires at least one PDU session"},
		{"too many PDU sessions", []string{"multi-ue", "-n", "1", "--numPduSessions=16"}, "--numPduSessions must be between 0 and 15"},
		{"negative duration", []string{"multi-ue", "-n", "1", "--timeBetweenRegistration=-1"}, "--timeBetweenRegistration cannot be negative"},
		{"negative loop count", []string{"multi-ue", "-n", "1", "--loopCount=-1"}, "--loopCount cannot be negative"},
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

func TestRegistrationOnlyHasNoPDUAndKeepsDefault(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		called := false
		app := parserApp(t, func(c *cli.Context) error {
			called = true
			want := 1
			if explicit {
				want = 0
			}
			if c.Int("numPduSessions") != want {
				t.Fatalf("sessions=%d, want %d", c.Int("numPduSessions"), want)
			}
			return nil
		})
		args := []string{"packetrusher", "multi-ue", "-n", "1"}
		if explicit {
			args = append(args, "--numPduSessions=0")
		}
		if err := app.Run(args); err != nil {
			t.Fatal(err)
		}
		if !called {
			t.Fatal("registration-only validation did not reach action")
		}
	}
}
