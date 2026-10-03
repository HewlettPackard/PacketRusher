// SPDX-License-Identifier: Apache-2.0
package main

import (
	"encoding/json"
	"fmt"
	"github.com/urfave/cli/v2"
	"my5G-RANTester/internal/buildinfo"
)

func versionCommand() *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "Print build version, source revision and toolchain",
		Flags: []cli.Flag{&cli.BoolFlag{Name: "json", Usage: "Print machine-readable build metadata"}},
		Action: func(c *cli.Context) error {
			info := buildinfo.Current()
			if c.Bool("json") {
				return json.NewEncoder(c.App.Writer).Encode(info)
			}
			_, err := fmt.Fprintf(c.App.Writer, "PacketRusher %s\nGo: %s %s/%s\n", info.String(), info.GoVersion, info.GOOS, info.GOARCH)
			return err
		},
	}
}
