/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2026 Valentin D'Emmanuele
 */
package main

import (
	"errors"
	"fmt"
	"my5G-RANTester/internal/analytics"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/urfave/cli/v2"
)

// withResults serves the procedure results on --metrics-addr while the action
// runs, and writes them to --report-json once it ends, even with an error.
func withResults(action cli.ActionFunc) cli.ActionFunc {
	return func(c *cli.Context) error {
		if address := c.String("metrics-addr"); address != "" {
			listener, err := net.Listen("tcp", address)
			if err != nil {
				return fmt.Errorf("listen for metrics on %s: %w", address, err)
			}
			defer listener.Close()
			mux := http.NewServeMux()
			mux.HandleFunc("/metrics", analytics.Metrics)
			go func() { _ = (&http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}).Serve(listener) }()
		}
		path := c.Path("report-json")
		if path == "" {
			return action(c)
		}
		// Before the run: a path that cannot be written would lose its results.
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		return errors.Join(action(c), analytics.WriteJSON(file), file.Close())
	}
}
