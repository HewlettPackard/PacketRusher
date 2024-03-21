/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2024 Hewlett Packard Enterprise Development LP
 */
package cmd

import (
	"fmt"
	"my5G-RANTester/config"
	"my5G-RANTester/internal/common/tools"
	"my5G-RANTester/internal/scenario"
	pcap "my5G-RANTester/internal/utils"
	"sort"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// singleUePduCmd represents the singleUePdu command
var singleUePduCmd = &cobra.Command{
	Use:   "single-ue-pdu",
	Short: "Load endurance stress tests.",
	Long: `Load endurance stress tests.
	This test case will launch one UE. See packetrusher single-ue-pdu --help,
	Example for testing single UE: single-ue`,
	Aliases: []string{"single"},
	Run: func(cmd *cobra.Command, args []string) {
		name := "Testing registration of single UE"
		cfgPath, _ := cmd.Flags().GetString("config")
		cfg := setConfig(cfgPath)

		log.Info("---------------------------------------")
		log.Info("[TESTER] Starting test function: ", name)
		log.Info("[TESTER][UE] Number of UEs: ", 1)
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

		if cmd.Flags().Lookup("pcap").Changed {
			pcap.CaptureTraffic(pcapPath)
		}
		if tunnel {
			vrf, _ := cmd.Flags().GetBool("tunnel-vrf")
			if vrf {
				tunnelMode = config.TunnelVrf
			} else {
				tunnelMode = config.TunnelTun
			}
		}

		timeBeforeDeregistration, _ := cmd.Flags().GetInt("timeBeforeDeregistration")
		timeBeforeNgapHandover, _ := cmd.Flags().GetInt("timeBeforeNgapHandover")
		timeBeforeXnHandover, _ := cmd.Flags().GetInt("timeBeforeXnHandover")

		loop, _ := cmd.Flags().GetBool("loop")
		timeBetweenRegistration, _ := cmd.Flags().GetInt("timeBetweenRegistration")
		numPduSessions, _ := cmd.Flags().GetInt("numPduSessions")
		timeBeforeIdle, _ := cmd.Flags().GetInt("timeBeforeIdle")

		testsingleUePdu(tunnelMode, loop, timeBetweenRegistration, timeBeforeDeregistration, timeBeforeNgapHandover, timeBeforeXnHandover, timeBeforeIdle, numPduSessions)

	},
}

func init() {
	rootCmd.AddCommand(singleUePduCmd)

	singleUePduCmd.Flags().IntP("timeBetweenRegistration", "R", 500, "The time in ms, between UE registration when looping.")
	singleUePduCmd.Flags().IntP("timeBeforeDeregistration", "D", 0, "The time in ms, before a UE deregisters once it has been registered. 0 to disable auto-deregistration.")
	singleUePduCmd.Flags().IntP("timeBeforeNgapHandover", "N", 0, "The time in ms, before triggering a UE handover using NGAP Handover. 0 to disable handover. This requires at least two gNodeB, eg: two N2/N3 IPs.")
	singleUePduCmd.Flags().IntP("timeBeforeXnHandover", "X", 0, "The time in ms, before triggering a UE handover using Xn Handover. 0 to disable handover. This requires at least two gNodeB, eg: two N2/N3 IPs.")
	singleUePduCmd.Flags().IntP("timeBeforeIdle", "I", 0, "The time in ms, before switching UE to Idle. 0 to disable Idling.")
	singleUePduCmd.Flags().IntP("timeBeforeServiceRequest", "S", 1000, "The time in ms, before reconnecting to gNodeB after switching to Idle state. Default is 1000 ms. Only work in conjunction with timeBeforeIdle.")
	singleUePduCmd.Flags().IntP("numPduSessions", "p", 1, "The number of PDU Sessions to create")
	singleUePduCmd.Flags().BoolP("loop", "l", false, "Register UE in a loop.")
	singleUePduCmd.Flags().BoolP("tunnel", "t", false, "Disable the creation of the GTP-U tunnel interface")
	singleUePduCmd.Flags().Bool("tunnel-vrf", true, "Enable/disable VRP usage of the GTP-U tunnel interface.")
	singleUePduCmd.Flags().String("pcap", "./dump.pcap", "Capture traffic to given PCAP file when a path is given")
}

func testsingleUePdu(tunnelMode config.TunnelMode, loop bool, timeBetweenRegistration int, timeBeforeDeregistration int, timeBeforeNgapHandover int, timeBeforeXnHandover int, timeBeforeIdle int, numPduSessions int) {
	if tunnelMode != config.TunnelDisabled && timeBetweenRegistration < 500 {
		log.Fatal("When using the --tunnel option, --timeBetweenRegistration must be equal to at least 500 ms, or else gtp5g kernel module may crash if you create tunnels too rapidly.")
	}

	if numPduSessions > 16 {
		log.Fatal("You can't have more than 16 PDU Sessions per UE as per spec.")
	}

	cfg := config.GetConfig()

	var err error
	gnb2 := cfg.GNodeB
	gnb2.PlmnList.GnbId = genereateGnbId(1, cfg.GNodeB.PlmnList.GnbId)

	gnb2.ControlIF.Ip, err = tools.IncrementIP(cfg.GNodeB.ControlIF.Ip, "0.0.0.0/0")
	if err != nil {
		log.Fatal("[GNB][CONFIG] Error while allocating ip for N2: " + err.Error())
	}
	gnb2.DataIF.Ip, err = tools.IncrementIP(cfg.GNodeB.DataIF.Ip, "0.0.0.0/0")
	if err != nil {
		log.Fatal("[GNB][CONFIG] Error while allocating ip for N3: " + err.Error())
	}

	gnbs := []config.GNodeB{cfg.GNodeB, gnb2}
	nextgnb := 0

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
	for i := 0; i < len(tasks); i++ {
		tasks[i].Delay = tasks[i].Delay - sumDelay
		sumDelay += tasks[i].Delay

		if tasks[i].TaskType == scenario.NGAPHandover || tasks[i].TaskType == scenario.XNHandover {
			nextgnb = (nextgnb + 1) % 2
			tasks[i].Parameters.GnbId = gnbs[nextgnb].PlmnList.GnbId
		}
	}

	ueScenario := scenario.UEScenario{
		Config: cfg.Ue,
		Tasks:  tasks,
	}
	ueScenarios := []scenario.UEScenario{ueScenario}

	if loop {
		ueScenario.Loop = timeBetweenRegistration
	}

	r := scenario.ScenarioManager{}
	r.Start(gnbs, cfg.AMF, ueScenarios, 0)
}

func genereateGnbId(i int, gnbId string) string {

	gnbId_int, err := strconv.ParseInt(gnbId, 16, 0)
	if err != nil {
		log.Fatal("[UE][CONFIG] Given gnbId is invalid")
	}
	base := int(gnbId_int) + i

	gnbId = fmt.Sprintf("%06x", base)
	return gnbId
}
