/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2024 Hewlett Packard Enterprise Development LP
 */
package cmd

import (
	"my5G-RANTester/internal/templates"
	pcap "my5G-RANTester/internal/utils"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// ue represents the ue command
var ue = &cobra.Command{
	Use:   "ue",
	Short: "Launch a gNB and a UE with a PDU Session\nFor more complex scenario and features, use instead packetrusher multi-ue\n",
	Run: func(cmd *cobra.Command, args []string) {
		name := "Testing an ue attached with configuration"
		cfgPath, _ := cmd.Flags().GetString("config")
		cfg := setConfig(cfgPath)
		tunnelEnabled, _ := cmd.Flags().GetBool("disableTunnel")

		log.Info("---------------------------------------")
		log.Info("[TESTER] Starting test function: ", name)
		log.Info("[TESTER][UE] Number of UEs: ", 1)
		log.Info("[TESTER][UE] disableTunnel is ", !tunnelEnabled)
		log.Info("[TESTER][GNB] Control interface IP/Port: ", cfg.GNodeB.ControlIF.Ip, "/", cfg.GNodeB.ControlIF.Port)
		log.Info("[TESTER][GNB] Data interface IP/Port: ", cfg.GNodeB.DataIF.Ip, "/", cfg.GNodeB.DataIF.Port)
		log.Info("[TESTER][AMF] AMF IP/Port: ", cfg.AMF.Ip, "/", cfg.AMF.Port)
		log.Info("---------------------------------------")

		pcapPath, _ := cmd.Flags().GetString("pcap")

		if cmd.Flags().Lookup("pcap").Changed {
			pcap.CaptureTraffic(pcapPath)
		}

		templates.TestAttachUeWithConfiguration(tunnelEnabled)
	},
}

func init() {
	rootCmd.AddCommand(ue)

	ue.Flags().BoolP("disableTunnel", "t", false, "Disable the creation of the GTP-U tunnel interface")
	ue.Flags().String("pcap", "./dump.pcap", "Capture traffic to given PCAP file when a path is given")
}
