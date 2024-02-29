package main

import (
	"my5G-RANTester/config"
	"my5G-RANTester/internal/templates"
	pcap "my5G-RANTester/internal/utils"

	// "fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const version = "1.0.1"

func init() {

	spew.Config.Indent = "\t"

}

func main() {

	app := &cli.App{
		Flags: []cli.Flag{
			&cli.PathFlag{Name: "config", Usage: "Configuration file path. (Default: ./config/config.yml)"},
		},
		Commands: []*cli.Command{
			{
				Name:    "single-ue-pdu",
				Aliases: []string{"single-ue"},
				Usage: "\nLoad endurance stress tests.\n" +
					"Example for testing UE: single-ue\n" +
					"This test case will launch one UE. See packetrusher single-ue --help\n",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "timeBetweenRegistration", Value: 500, Aliases: []string{"tr"}, Usage: "The time in ms, between UE registration."},
					&cli.IntFlag{Name: "timeBeforeDeregistration", Value: 0, Aliases: []string{"td"}, Usage: "The time in ms, before UE deregisters once it has been registered. 0 to disable auto-deregistration."},
					&cli.IntFlag{Name: "timeBeforeNgapHandover", Value: 0, Aliases: []string{"ngh"}, Usage: "The time in ms, before triggering a UE handover using NGAP Handover. 0 to disable handover. This requires at least two gNodeB, eg: two N2/N3 IPs."},
					&cli.IntFlag{Name: "timeBeforeXnHandover", Value: 0, Aliases: []string{"xnh"}, Usage: "The time in ms, before triggering a UE handover using Xn Handover. 0 to disable handover. This requires at least two gNodeB, eg: two N2/N3 IPs."},
					&cli.IntFlag{Name: "timeBeforeIdle", Value: 0, Aliases: []string{"idl"}, Usage: "The time in ms, before switching UE to Idle. 0 to disable Idling."},
					&cli.IntFlag{Name: "numPduSessions", Value: 1, Aliases: []string{"nPdu"}, Usage: "The number of PDU Sessions to create"},
					&cli.BoolFlag{Name: "loop", Aliases: []string{"l"}, Usage: "Register UE in a loop."},
					&cli.BoolFlag{Name: "tunnel", Aliases: []string{"t"}, Usage: "Enable the creation of the GTP-U tunnel interface."},
					&cli.BoolFlag{Name: "tunnel-vrf", Value: true, Usage: "Enable/disable VRP usage of the GTP-U tunnel interface."},
					&cli.PathFlag{Name: "pcap", Usage: "Capture traffic to given PCAP file when a path is given", Value: "./dump.pcap"},
				},
				Action: func(c *cli.Context) error {
					var numUes int
					name := "Testing registration of single UE"
					cfg := setConfig(*c)

					log.Info("PacketRusher version " + version)
					log.Info("---------------------------------------")
					log.Info("[TESTER] Starting test function: ", name)
					log.Info("[TESTER][UE] Number of UEs: ", numUes)
					log.Info("[TESTER][GNB] gNodeB control interface IP/Port: ", cfg.GNodeB.ControlIF.Ip, "/", cfg.GNodeB.ControlIF.Port)
					log.Info("[TESTER][GNB] gNodeB data interface IP/Port: ", cfg.GNodeB.DataIF.Ip, "/", cfg.GNodeB.DataIF.Port)
					log.Info("[TESTER][AMF] AMF IP/Port: ", cfg.AMF.Ip, "/", cfg.AMF.Port)
					log.Info("---------------------------------------")

					if c.IsSet("pcap") {
						pcap.CaptureTraffic(c.Path("pcap"))
					}

					tunnelMode := config.TunnelDisabled
					if c.Bool("tunnel") {
						if c.Bool("tunnel-vrf") {
							tunnelMode = config.TunnelVrf
						} else {
							tunnelMode = config.TunnelTun
						}
					}
					templates.TestMultiUesInQueue(1, tunnelMode, true, c.Bool("loop"), c.Int("timeBetweenRegistration"), c.Int("timeBeforeDeregistration"), c.Int("timeBeforeNgapHandover"), c.Int("timeBeforeXnHandover"), c.Int("timeBeforeIdle"), c.Int("numPduSessions"), 0)

					return nil
				},
			},
			{
				Name:    "multi-ue-pdu",
				Aliases: []string{"multi-ue"},
				Usage: "\nLoad endurance stress tests.\n" +
					"Example for testing multiple UEs: multi-ue -n 5 \n" +
					"This test case will launch N UEs. See packetrusher multi-ue --help\n",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "number-of-ues", Value: 1, Aliases: []string{"n"}},
					&cli.IntFlag{Name: "timeBetweenRegistration", Value: 500, Aliases: []string{"tr"}, Usage: "The time in ms, between UE registration."},
					&cli.IntFlag{Name: "timeBeforeDeregistration", Value: 0, Aliases: []string{"td"}, Usage: "The time in ms, before a UE deregisters once it has been registered. 0 to disable auto-deregistration."},
					&cli.IntFlag{Name: "timeBeforeNgapHandover", Value: 0, Aliases: []string{"ngh"}, Usage: "The time in ms, before triggering a UE handover using NGAP Handover. 0 to disable handover. This requires at least two gNodeB, eg: two N2/N3 IPs."},
					&cli.IntFlag{Name: "timeBeforeXnHandover", Value: 0, Aliases: []string{"xnh"}, Usage: "The time in ms, before triggering a UE handover using Xn Handover. 0 to disable handover. This requires at least two gNodeB, eg: two N2/N3 IPs."},
					&cli.IntFlag{Name: "timeBeforeIdle", Value: 0, Aliases: []string{"idl"}, Usage: "The time in ms, before switching UE to Idle. 0 to disable Idling."},
					&cli.IntFlag{Name: "timeBeforeReconnecting", Value: 1000, Aliases: []string{"tbr"}, Usage: "The time in ms, before reconnecting to gNodeB after switching to Idle state. Default is 1000 ms. Only work in conjunction with timeBeforeIdle."},
					&cli.IntFlag{Name: "numPduSessions", Value: 1, Aliases: []string{"nPdu"}, Usage: "The number of PDU Sessions to create"},
					&cli.BoolFlag{Name: "loop", Aliases: []string{"l"}, Usage: "Register UEs in a loop."},
					&cli.BoolFlag{Name: "tunnel", Aliases: []string{"t"}, Usage: "Enable the creation of the GTP-U tunnel interface."},
					&cli.BoolFlag{Name: "tunnel-vrf", Value: true, Usage: "Enable/disable VRP usage of the GTP-U tunnel interface."},
					&cli.BoolFlag{Name: "dedicatedGnb", Aliases: []string{"d"}, Usage: "Enable the creation of a dedicated gNB per UE. Require one IP on N2/N3 per gNB."},
					&cli.IntFlag{Name: "requestRate", Value: 0, Aliases: []string{"rate"}, Usage: "The number of max requests sent by seconds. 0 for unlimited"},
					&cli.PathFlag{Name: "pcap", Usage: "Capture traffic to given PCAP file when a path is given", Value: "./dump.pcap"},
				},
				Action: func(c *cli.Context) error {
					var numUes int
					name := "Testing registration of multiple UEs"
					cfg := setConfig(*c)
					if c.IsSet("number-of-ues") {
						numUes = c.Int("number-of-ues")
					} else {
						log.Info(c.Command.Usage)
						return nil
					}

					log.Info("PacketRusher version " + version)
					log.Info("---------------------------------------")
					log.Info("[TESTER] Starting test function: ", name)
					log.Info("[TESTER][UE] Number of UEs: ", numUes)
					log.Info("[TESTER][GNB] gNodeB control interface IP/Port: ", cfg.GNodeB.ControlIF.Ip, "/", cfg.GNodeB.ControlIF.Port)
					log.Info("[TESTER][GNB] gNodeB data interface IP/Port: ", cfg.GNodeB.DataIF.Ip, "/", cfg.GNodeB.DataIF.Port)
					log.Info("[TESTER][AMF] AMF IP/Port: ", cfg.AMF.Ip, "/", cfg.AMF.Port)
					log.Info("---------------------------------------")

					if c.IsSet("pcap") {
						pcap.CaptureTraffic(c.Path("pcap"))
					}

					tunnelMode := config.TunnelDisabled
					if c.Bool("tunnel") {
						if c.Bool("tunnel-vrf") {
							tunnelMode = config.TunnelVrf
						} else {
							tunnelMode = config.TunnelTun
						}
					}
					templates.TestMultiUesInQueue(numUes, tunnelMode, c.Bool("dedicatedGnb"), c.Bool("loop"), c.Int("timeBetweenRegistration"), c.Int("timeBeforeDeregistration"), c.Int("timeBeforeNgapHandover"), c.Int("timeBeforeXnHandover"), c.Int("timeBeforeIdle"), c.Int("timeBeforeReconnecting"), c.Int("numPduSessions"), c.Int("rate"))

					return nil
				},
			},
			{
				Name:    "custom-scenario",
				Aliases: []string{"c"},
				Flags: []cli.Flag{
					&cli.PathFlag{Name: "scenario", Usage: "Specify the scenario path in .wasm"},
				},
				Action: func(c *cli.Context) error {
					setConfig(*c)

					var scenarioPath string

					if c.IsSet("scenario") {
						scenarioPath = c.Path("scenario")
					} else {
						log.Info(c.Command.Usage)
						return nil
					}

					templates.TestWithCustomScenario(scenarioPath)

					return nil
				},
			},
			{
				Name:    "amf-load-loop",
				Aliases: []string{"amf-load-loop"},
				Usage: "\nTest AMF responses in interval\n" +
					"Example for generating 20 requests to AMF per second in interval of 20 seconds: amf-load-loop -n 20 -t 20\n",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "number-of-requests", Value: 1, Aliases: []string{"n"}},
					&cli.IntFlag{Name: "time", Value: 1, Aliases: []string{"t"}},
				},
				Action: func(c *cli.Context) error {
					var time int
					var numRqs int

					name := "Test AMF responses in interval"
					cfg := setConfig(*c)

					numRqs = c.Int("number-of-requests")
					time = c.Int("time")

					log.Info("PacketRusher version " + version)
					log.Info("---------------------------------------")
					log.Warn("[TESTER] Starting test function: ", name)
					log.Warn("[TESTER][UE] Number of Requests per second: ", numRqs)
					log.Info("[TESTER][GNB] gNodeB control interface IP/Port: ", cfg.GNodeB.ControlIF.Ip, "/", cfg.GNodeB.ControlIF.Port)
					log.Info("[TESTER][GNB] gNodeB data interface IP/Port: ", cfg.GNodeB.DataIF.Ip, "/", cfg.GNodeB.DataIF.Port)
					log.Info("[TESTER][AMF] AMF IP/Port: ", cfg.AMF.Ip, "/", cfg.AMF.Port)
					log.Info("---------------------------------------")
					log.Warn("[TESTER][GNB] Total of AMF Responses in the interval:", templates.TestRqsLoop(numRqs, time))
					return nil
				},
			},
			{
				Name:    "Test availability of AMF",
				Aliases: []string{"amf-availability"},
				Usage: "\nTest availability of AMF in interval\n" +
					"Test availability of AMF in 20 seconds: amf-availability -t 20\n",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "time", Value: 1, Aliases: []string{"t"}},
				},
				Action: func(c *cli.Context) error {
					var time int

					name := "Test availability of AMF"
					cfg := setConfig(*c)
					time = c.Int("time")

					log.Info("PacketRusher version " + version)
					log.Info("---------------------------------------")
					log.Warn("[TESTER] Starting test function: ", name)
					log.Warn("[TESTER][UE] Interval of test: ", time, " seconds")
					log.Info("[TESTER][GNB] Control interface IP/Port: ", cfg.GNodeB.ControlIF.Ip, "/", cfg.GNodeB.ControlIF.Port)
					log.Info("[TESTER][GNB] Data interface IP/Port: ", cfg.GNodeB.DataIF.Ip, "/", cfg.GNodeB.DataIF.Port)
					log.Info("[TESTER][AMF] AMF IP/Port: ", cfg.AMF.Ip, "/", cfg.AMF.Port)
					log.Info("---------------------------------------")
					templates.TestAvailability(time)
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

func setConfig(c cli.Context) config.Config {
	var cfg config.Config
	if c.IsSet("config") {
		cfg = config.Load(c.Path("config"))
	} else {
		cfg = config.LoadDefaultConfig()
	}
	return cfg
}
