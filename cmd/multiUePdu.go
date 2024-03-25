/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2024 Hewlett Packard Enterprise Development LP
 */
package cmd

import (
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/tools"
	"my5G-RANTester/internal/scenario"
	pcap "my5G-RANTester/internal/utils"
	"sort"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// multiUePduCmd represents the multiUePdu command
var multiUePduCmd = &cobra.Command{
	Use:   "multi-ue-pdu",
	Short: "Load endurance stress tests.",
	Long: `Load endurance stress tests.
	This test case will launch N UEs. See packetrusher multi-ue --help,
	Example for testing multiple UEs: multi-ue -n 5`,
	Aliases: []string{"multi-ue"},
	Run: func(cmd *cobra.Command, args []string) {
		numUes, _ := cmd.Flags().GetInt("number-of-ues")
		name := "Testing registration of multiple UEs"
		cfgPath, _ := cmd.Flags().GetString("config")
		cfg := setConfig(cfgPath)

		log.Info("PacketRusher version " + version)
		log.Info("---------------------------------------")
		log.Info("[TESTER] Starting test function: ", name)
		log.Info("[TESTER][UE] Number of UEs: ", numUes)
		log.Info("[TESTER][GNB] gNodeB control interface IP/Port: ", cfg.GNodeB.ControlIF.Ip, "/", cfg.GNodeB.ControlIF.Port)
		log.Info("[TESTER][GNB] gNodeB data interface IP/Port: ", cfg.GNodeB.DataIF.Ip, "/", cfg.GNodeB.DataIF.Port)
		log.Info("[TESTER][AMF] AMF IP/Port: ", cfg.AMF.Ip, "/", cfg.AMF.Port)
		log.Info("---------------------------------------")

		pcapPath, _ := cmd.Flags().GetString("pcap")
		if cmd.Flags().Lookup("pcap").Changed {
			pcap.CaptureTraffic(pcapPath)
		}

		tunnelMode := config.TunnelDisabled

		tunnel, _ := cmd.Flags().GetBool("tunnel")

		if tunnel {
			vrf, _ := cmd.Flags().GetBool("tunnel-vrf")
			if vrf {
				tunnelMode = config.TunnelVrf
			} else {
				tunnelMode = config.TunnelTun
			}
		}

		dedicatedGnb, _ := cmd.Flags().GetBool("dedicatedGnb")
		loop, _ := cmd.Flags().GetBool("loop")
		timeBetweenRegistration, _ := cmd.Flags().GetInt("timeBetweenRegistration")
		timeBeforeDeregistration, _ := cmd.Flags().GetInt("timeBeforeDeregistration")
		timeBeforeNgapHandover, _ := cmd.Flags().GetInt("timeBeforeNgapHandover")
		timeBeforeXnHandover, _ := cmd.Flags().GetInt("timeBeforeXnHandover")
		numPduSessions, _ := cmd.Flags().GetInt("numPduSessions")
		timeBeforeIdle, _ := cmd.Flags().GetInt("timeBeforeIdle")

		timeBeforeServiceRequest, _ := cmd.Flags().GetInt("timeBeforeServiceRequest")
		maxRequestRate, _ := cmd.Flags().GetInt("maxRequestRate")

		testMultiUesInQueue(numUes, tunnelMode, dedicatedGnb, loop, timeBetweenRegistration, timeBeforeDeregistration, timeBeforeNgapHandover, timeBeforeXnHandover, timeBeforeIdle, timeBeforeServiceRequest, numPduSessions, maxRequestRate)

	},
}

func init() {
	rootCmd.AddCommand(multiUePduCmd)

	multiUePduCmd.Flags().IntP("number-of-ues", "n", 1, "Number of UE to be created")
	multiUePduCmd.Flags().IntP("timeBetweenRegistration", "R", 500, "The time in ms, between UE registration.")
	multiUePduCmd.Flags().IntP("timeBeforeDeregistration", "D", 0, "The time in ms, before a UE deregisters once it has been registered. 0 to disable auto-deregistration.")
	multiUePduCmd.Flags().IntP("timeBeforeNgapHandover", "N", 0, "The time in ms, before triggering a UE handover using NGAP Handover. 0 to disable handover. This requires at least two gNodeB, eg: two N2/N3 IPs.")
	multiUePduCmd.Flags().IntP("timeBeforeXnHandover", "X", 0, "The time in ms, before triggering a UE handover using Xn Handover. 0 to disable handover. This requires at least two gNodeB, eg: two N2/N3 IPs.")
	multiUePduCmd.Flags().IntP("timeBeforeIdle", "I", 0, "The time in ms, before switching UE to Idle. 0 to disable Idling.")
	multiUePduCmd.Flags().IntP("numPduSessions", "p", 1, "The number of PDU Sessions to create")
	multiUePduCmd.Flags().IntP("maxRequestRate", "r", 0, "Max number of messages send to the AMF per seconds. This takes priority over others \"timeBefore\" flags. 0 to disable.")
	multiUePduCmd.Flags().BoolP("loop", "l", false, "")
	multiUePduCmd.Flags().BoolP("tunnel", "t", false, "Disable the creation of the GTP-U tunnel interface")
	multiUePduCmd.Flags().Bool("tunnel-vrf", false, "Disable the creation of the GTP-U tunnel interface")
	multiUePduCmd.Flags().BoolP("dedicatedGnb", "d", false, "Disable the creation of the GTP-U tunnel interface")
	multiUePduCmd.Flags().Bool("pcap", false, "Disable the creation of the GTP-U tunnel interface")
	multiUePduCmd.Flags().IntP("timeBeforeServiceRequest", "S", 1000, "The time in ms, before reconnecting to gNodeB after switching to Idle state. Default is 1000 ms. Only work in conjunction with timeBeforeIdle.")
}

func testMultiUesInQueue(numUes int, tunnelMode config.TunnelMode, dedicatedGnb bool, loop bool, timeBetweenRegistration int, timeBeforeDeregistration int, timeBeforeNgapHandover int, timeBeforeXnHandover int, timeBeforeIdle int, timeBeforeServiceRequest int, numPduSessions int, maxRequestRate int) {
	if tunnelMode != config.TunnelDisabled && timeBetweenRegistration < 500 {
		log.Fatal("When using the --tunnel option, --timeBetweenRegistration must be equal to at least 500 ms, or else gtp5g kernel module may crash if you create tunnels too rapidly.")
	}

	if numPduSessions > 16 {
		log.Fatal("You can't have more than 16 PDU Sessions per UE as per spec.")
	}

	cfg := config.GetConfig()
	cfg.Ue.TunnelMode = tunnelMode

	gnb := cfg.GNodeB
	gnbs := []config.GNodeB{gnb}
	nextgnb := 0
	if timeBeforeNgapHandover > 0 || timeBeforeXnHandover > 0 {
		gnb = nextGnbConf(gnb, 1, cfg.GNodeB.PlmnList.GnbId)
		gnbs = append(gnbs, gnb)
	}
	ueScenarios := []scenario.UEScenario{}
	for i := 0; i < numUes; i++ {
		if dedicatedGnb && i != 0 {
			gnb = nextGnbConf(gnb, len(gnbs), cfg.GNodeB.PlmnList.GnbId)
			gnbs = append(gnbs, gnb)
			nextgnb = len(gnbs) - 1
			if timeBeforeNgapHandover > 0 || timeBeforeXnHandover > 0 {
				gnb = nextGnbConf(gnb, len(gnbs), cfg.GNodeB.PlmnList.GnbId)
				gnbs = append(gnbs, gnb)
			}
		}

		tasks := []scenario.Task{
			{
				TaskType: scenario.AttachToGNB,
				Parameters: struct {
					GnbId string
				}{gnbs[nextgnb].PlmnList.GnbId},
			},
			{
				TaskType: scenario.Registration,
			},
		}

		for i := 0; i < numPduSessions; i++ {
			tasks = append(tasks, scenario.Task{
				TaskType: scenario.NewPDUSession,
			})
		}

		if timeBeforeNgapHandover != 0 {
			tasks = append(tasks, scenario.Task{
				TaskType: scenario.NGAPHandover,
				Delay:    timeBeforeNgapHandover,
			})
		}

		if timeBeforeXnHandover != 0 {
			tasks = append(tasks, scenario.Task{
				TaskType: scenario.XNHandover,
				Delay:    timeBeforeXnHandover,
			})
		}
		if timeBeforeIdle != 0 {
			tasks = append(tasks, scenario.Task{
				TaskType: scenario.Idle,
				Delay:    timeBeforeIdle,
			})
			if timeBeforeServiceRequest != 0 {
				tasks = append(tasks, scenario.Task{
					TaskType: scenario.ServiceRequest,
					Delay:    timeBeforeIdle + timeBeforeServiceRequest,
				})
			}
		}

		if timeBeforeDeregistration != 0 {
			tasks = append(tasks, scenario.Task{
				TaskType: scenario.Deregistration,
				Delay:    timeBeforeDeregistration,
			})
		}

		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].Delay < tasks[j].Delay
		})

		sumDelay := 0
		for j := 2; j < len(tasks); j++ {
			tasks[j].Delay = max(0, tasks[j].Delay-sumDelay)
			sumDelay += tasks[j].Delay

			if tasks[j].TaskType == scenario.NGAPHandover || tasks[j].TaskType == scenario.XNHandover {
				tasks[j].Parameters.GnbId = gnbs[(len(gnbs)-2)+(nextgnb+1)%2].PlmnList.GnbId
			}
		}

		tasks[1].Delay = i * timeBetweenRegistration

		ueCfg := cfg.Ue
		ueCfg.Msin = tools.IncrementMsin(i+1, cfg.Ue.Msin)
		ueScenario := scenario.UEScenario{
			Config: ueCfg,
			Tasks:  tasks,
		}

		if loop {
			ueScenario.Loop = timeBetweenRegistration // TODO: change this!
		}

		ueScenarios = append(ueScenarios, ueScenario)
	}

	r := scenario.ScenarioManager{}
	r.Start(gnbs, cfg.AMF, ueScenarios, maxRequestRate)
}

func nextGnbConf(gnb config.GNodeB, i int, baseId string) config.GNodeB {
	var err error
	gnb.PlmnList.GnbId = genereateGnbId(i, baseId)
	gnb.ControlIF.Ip, err = tools.IncrementIP(gnb.ControlIF.Ip, "0.0.0.0/0")
	if err != nil {
		log.Fatal("[GNB][CONFIG] Error while allocating ip for N2: " + err.Error())
	}
	gnb.DataIF.Ip, err = tools.IncrementIP(gnb.DataIF.Ip, "0.0.0.0/0")
	if err != nil {
		log.Fatal("[GNB][CONFIG] Error while allocating ip for N3: " + err.Error())
	}
	return gnb
}
