package main

import (
	"my5G-RANTester/config"
	"my5G-RANTester/internal/templates"

	// "fmt"
	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
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

func main() {

	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "ue",
				Aliases: []string{"ue"},
				Usage:   "Testing an ue attached with configuration",
				Action: func(c *cli.Context) error {
					name := "Testing an ue attached with configuration"
					cfg := config.Data

					log.Info("---------------------------------------")
					log.Info("[TESTER] Starting test function: ", name)
					log.Info("[TESTER][UE] Number of UEs: ", 1)
					log.Info("[TESTER][GNB] Control interface IP/Port: ", cfg.GNodeB.ControlIF.Ip, "/", cfg.GNodeB.ControlIF.Port)
					log.Info("[TESTER][GNB] Data interface IP/Port: ", cfg.GNodeB.DataIF.Ip, "/", cfg.GNodeB.DataIF.Port)
					log.Info("[TESTER][AMF] AMF IP/Port: ", cfg.AMF.Ip, "/", cfg.AMF.Port)
					log.Info("---------------------------------------")
					templates.TestAttachUeWithConfiguration()
					return nil
				},
			},
			{
				Name:    "load-test",
				Aliases: []string{"load-test"},
				Usage: "\nLoad endurance stress tests.\n" +
					"Example for testing multiple UEs: load-test -n 5 \n",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "number-of-ues", Value: 1, Aliases: []string{"n"}},
				},
				Action: func(c *cli.Context) error {
					var numUes int
					name := "Testing registration of multiple UEs"
					cfg := config.Data

					if c.IsSet("number-of-ues") {
						numUes = c.Int("number-of-ues")
					} else {
						log.Info(c.Command.Usage)
						return nil
					}

					log.Info("---------------------------------------")
					log.Info("[TESTER] Starting test function: ", name)
					log.Info("[TESTER][UE] Number of UEs: ", numUes)
					log.Info("[TESTER][GNB] gNodeB control interface IP/Port: ", cfg.GNodeB.ControlIF.Ip, "/", cfg.GNodeB.ControlIF.Port)
					log.Info("[TESTER][GNB] gNodeB data interface IP/Port: ", cfg.GNodeB.DataIF.Ip, "/", cfg.GNodeB.DataIF.Port)
					log.Info("[TESTER][AMF] AMF IP/Port: ", cfg.AMF.Ip, "/", cfg.AMF.Port)
					log.Info("---------------------------------------")
					templates.TestMultiUesInQueue(numUes)

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
