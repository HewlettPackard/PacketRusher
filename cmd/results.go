// SPDX-License-Identifier: Apache-2.0
package main

import (
	"context"
	"errors"
	"fmt"
	"my5G-RANTester/internal/analytics"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/urfave/cli/v2"
)

// Initialize reports only after this command's validation succeeds. App.Before
// runs before command flags are parsed and can reserve files for rejected input.
func resultsBefore(validate, start cli.BeforeFunc) cli.BeforeFunc {
	return func(c *cli.Context) error {
		if validate != nil {
			if err := validate(c); err != nil {
				return err
			}
		}
		return start(c)
	}
}

func resultsHooks() (cli.BeforeFunc, cli.AfterFunc) {
	var recorder *analytics.Recorder
	var server *http.Server
	var jsonFile, csvFile *os.File
	before := func(c *cli.Context) error {
		jsonPath, csvPath, address := c.Path("report-json"), c.Path("report-csv"), c.String("metrics-addr")
		if jsonPath == "" && csvPath == "" && address == "" {
			return nil
		}
		if jsonPath != "" && csvPath != "" && filepath.Clean(jsonPath) == filepath.Clean(csvPath) {
			return fmt.Errorf("JSON and CSV reports must use different paths")
		}
		// Reserve output names before starting a load test; existing files are kept.
		create := func(path string) (*os.File, error) {
			if path == "" {
				return nil, nil
			}
			file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
			if err != nil {
				return nil, fmt.Errorf("create report %s: %w", path, err)
			}
			return file, nil
		}
		var err error
		jsonFile, err = create(jsonPath)
		if err != nil {
			return err
		}
		csvFile, err = create(csvPath)
		if err != nil {
			if jsonFile != nil {
				jsonFile.Close()
				os.Remove(jsonFile.Name())
				jsonFile = nil
			}
			return err
		}
		recorder = analytics.NewRecorder()
		if address != "" {
			listener, err := net.Listen("tcp", address)
			if err != nil {
				for _, file := range []*os.File{jsonFile, csvFile} {
					if file != nil {
						file.Close()
						os.Remove(file.Name())
					}
				}
				jsonFile, csvFile, recorder = nil, nil, nil
				return fmt.Errorf("listen for metrics on %s: %w", address, err)
			}
			server = &http.Server{Handler: recorder, ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 30 * time.Second}
			go func() { _ = server.Serve(listener) }()
		}
		analytics.SetCurrent(recorder)
		return nil
	}
	after := func(*cli.Context) error {
		if recorder == nil {
			return nil
		}
		recorder.Close()
		analytics.SetCurrent(nil)
		var errs []error
		if server != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			errs = append(errs, server.Shutdown(ctx))
			cancel()
		}
		if jsonFile != nil {
			errs = append(errs, recorder.WriteJSON(jsonFile), jsonFile.Close())
		}
		if csvFile != nil {
			errs = append(errs, recorder.WriteCSV(csvFile), csvFile.Close())
		}
		recorder = nil
		return errors.Join(errs...)
	}
	return before, after
}
