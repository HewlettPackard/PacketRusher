/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2024 Hewlett Packard Enterprise Development LP
 */
package cmd

import (
	"my5G-RANTester/config"
	pcap "my5G-RANTester/internal/utils"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// singleUePduCmd represents the singleUePdu command
var singleUePduCmd = &cobra.Command{
	Use:   "single-ue",
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
		timeBeforeServiceRequest, _ := cmd.Flags().GetInt("timeBeforeServiceRequest")

		testMultiUesInQueue(1, tunnelMode, false, loop, timeBetweenRegistration, timeBeforeDeregistration, timeBeforeNgapHandover, timeBeforeXnHandover, timeBeforeIdle, timeBeforeServiceRequest, numPduSessions, 0)

	},
}

func init() {
	rootCmd.AddCommand(singleUePduCmd)

	singleUePduCmd.Flags().IntP("timeBetweenRegistration", "R", 500, "The time in ms, between UE registration when looping.")
	singleUePduCmd.Flags().IntP("timeBeforeDeregistration", "D", 0, "The time in ms, before a UE deregisters once it has been registered. 0 to disable auto-deregistration.")
	singleUePduCmd.Flags().IntP("timeBeforeNgapHandover", "N", 0, "The time in ms, before triggering a UE handover using NGAP Handover. 0 to disable handover. This requires at least two gNodeB, eg: two N2/N3 IPs.")
	singleUePduCmd.Flags().IntP("timeBeforeXnHandover", "X", 0, "The time in ms, before triggering a UE handover using Xn Handover. 0 to disable handover. This requires at least two gNodeB, eg: two N2/N3 IPs.")
	singleUePduCmd.Flags().IntP("timeBeforeIdle", "I", 0, "The time in ms, before switching UE to Idle. 0 to disable Idling.")
	singleUePduCmd.Flags().IntP("timeBeforeServiceRequest", "S", 10000, "The time in ms, before reconnecting to gNodeB after switching to Idle state. Default is 1000 ms. Only work in conjunction with timeBeforeIdle.")
	singleUePduCmd.Flags().IntP("numPduSessions", "p", 1, "The number of PDU Sessions to create")
	singleUePduCmd.Flags().BoolP("loop", "l", false, "Register UE in a loop.")
	singleUePduCmd.Flags().BoolP("tunnel", "t", false, "Disable the creation of the GTP-U tunnel interface")
	singleUePduCmd.Flags().Bool("tunnel-vrf", true, "Enable/disable VRP usage of the GTP-U tunnel interface.")
	singleUePduCmd.Flags().String("pcap", "./dump.pcap", "Capture traffic to given PCAP file when a path is given")
}
