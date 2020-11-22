package main

import (
	"my5G-RANTester/config"
	// "fmt"
	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"my5G-RANTester/internal/templates"
	"os"
)

const version = "0.1"

func init() {
	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)
	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)
	spew.Config.Indent = "\t"

	log.Info("my5G-RANTester version " + version)

}

func execLoadTest(name string, numberUes int) {
	switch name {
	case "tnla":
		log.Debug(templates.TestMultiAttachUesInConcurrencyWithTNLAs(numberUes))
	case "gnb":
		log.Debug(templates.TestMultiAttachUesInConcurrencyWithGNBs(numberUes))
	default:
		log.Debug(templates.TestMultiAttachUesInQueue(numberUes))
	}
}

func main() {

	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "load-test",
				Aliases: []string{"load-test"},
				Usage: "\nLoad endurance stress tests.\n" +
					"Example: load-test -n 5 \n" +
					"Example for concurrency testing with different GNBs: load-test -g -n 10\n" +
					"Example for concurrency testing with some TNLAs: load-test -t -n 10\n",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "number-of-ues", Value: 1, Aliases: []string{"n"}},
					&cli.BoolFlag{Name: "gnb", Aliases: []string{"g"}},
					&cli.BoolFlag{Name: "tnla", Aliases: []string{"t"}},
				},
				Action: func(c *cli.Context) error {
					numUes := 1
					execName := "queue"
					name := "Multi attach UEs in queue"
					cfg := config.Data

					if c.IsSet("number-of-ues") {
						numUes = c.Int("number-of-ues")
					}

					if c.Bool("tnla") {
						execName = "tnla"
						name = "Multi attach UEs in concurrency with TNLAs"
					} else if c.Bool("gnb") {
						execName = "gnb"
						name = "Multi attach UEs in concurrency with GNBs"
					}
					log.Info("---------------------------------------")
					log.Info("Starting test function: ", name)
					log.Info("Number of UEs: ", numUes)
					log.Info("gNodeB control interface IP/Port: ", cfg.GNodeB.ControlIF.Ip, "/", cfg.GNodeB.ControlIF.Port)
					log.Info("gNodeB data interface IP/Port: ", cfg.GNodeB.DataIF.Ip, "/", cfg.GNodeB.DataIF.Port)
					log.Info("AMF IP/Port: ", cfg.AMF.Ip, "/", cfg.AMF.Port)
					log.Info("UPF IP/Port: ", cfg.UPF.Ip, "/", cfg.UPF.Port)
					log.Info("---------------------------------------")
					execLoadTest(execName, numUes)

					return nil
				},
			},
			{
				Name:    "ue",
				Aliases: []string{"ue"},
				Usage:   "test attach for ue with configuration",
				Action: func(c *cli.Context) error {
					log.Info(templates.TestAttachUeWithConfiguration())
					return nil
				},
			},
			{
				Name:    "gnb",
				Aliases: []string{"gnb"},
				Usage: "test attach with some gnbs with configuration.\n" +
					"Example for testing attached gnbs: gnb -n 5",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "number-of-gnbs", Value: 1, Aliases: []string{"n"}},
				},
				Action: func(c *cli.Context) error {
					numGnbs := c.Int("number-of-gnbs")

					log.Debug(templates.TestMultiAttachGnbInConcurrency(numGnbs))
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
