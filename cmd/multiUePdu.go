/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2024 Hewlett Packard Enterprise Development LP
 */
package cmd

import (
	"my5G-RANTester/config"
	"my5G-RANTester/internal/templates"
	pcap "my5G-RANTester/internal/utils"

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

		dedicatedGnb, _ := cmd.Flags().GetBool("dedicatedGnb")
		loop, _ := cmd.Flags().GetBool("loop")
		timeBetweenRegistration, _ := cmd.Flags().GetInt("timeBetweenRegistration")
		timeBeforeDeregistration, _ := cmd.Flags().GetInt("timeBeforeDeregistration")
		timeBeforeNgapHandover, _ := cmd.Flags().GetInt("timeBeforeNgapHandover")
		timeBeforeXnHandover, _ := cmd.Flags().GetInt("timeBeforeXnHandover")
		numPduSessions, _ := cmd.Flags().GetInt("numPduSessions")
		timeBeforeIdle, _ := cmd.Flags().GetInt("timeBeforeIdle")
		timeBeforeServiceRequest, _ := cmd.Flags().GetInt("timeBeforeServiceRequest")

		templates.TestMultiUesInQueue(numUes, tunnelMode, dedicatedGnb, loop, timeBetweenRegistration, timeBeforeDeregistration, timeBeforeNgapHandover, timeBeforeXnHandover, timeBeforeIdle, timeBeforeServiceRequest, numPduSessions, 0)

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
	multiUePduCmd.Flags().BoolP("loop", "l", false, "")
	multiUePduCmd.Flags().BoolP("tunnel", "t", false, "Disable the creation of the GTP-U tunnel interface")
	multiUePduCmd.Flags().Bool("tunnel-vrf", false, "Disable the creation of the GTP-U tunnel interface")
	multiUePduCmd.Flags().BoolP("dedicatedGnb", "d", false, "Disable the creation of the GTP-U tunnel interface")
	multiUePduCmd.Flags().Bool("pcap", false, "Disable the creation of the GTP-U tunnel interface")
	singleUePduCmd.Flags().IntP("timeBeforeServiceRequest", "S", 1000, "The time in ms, before reconnecting to gNodeB after switching to Idle state. Default is 1000 ms. Only work in conjunction with timeBeforeIdle.")
}
