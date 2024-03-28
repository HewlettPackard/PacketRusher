/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2024 Hewlett Packard Enterprise Development LP
 */
package cmd

import (
	"my5G-RANTester/internal/templates"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// gnbCmd represents the gnb command
var gnbCmd = &cobra.Command{
	Use:   "gnb",
	Short: "Launch only a gNB",
	Run: func(cmd *cobra.Command, args []string) {
		name := "Testing an gnb attached with configuration"
		cfgPath, _ := cmd.Flags().GetString("config")
		cfg := setConfig(cfgPath)

		log.Info("PacketRusher version " + version)
		log.Info("---------------------------------------")
		log.Info("[TESTER] Starting test function: ", name)
		log.Info("[TESTER][GNB] Number of GNBs: ", 1)
		log.Info("[TESTER][GNB] Control interface IP/Port: ", cfg.GNodeB.ControlIF.Ip, "/", cfg.GNodeB.ControlIF.Port)
		log.Info("[TESTER][GNB] Data interface IP/Port: ", cfg.GNodeB.DataIF.Ip, "/", cfg.GNodeB.DataIF.Port)
		log.Info("[TESTER][AMF] AMF IP/Port: ", cfg.AMF.Ip, "/", cfg.AMF.Port)
		log.Info("---------------------------------------")
		templates.TestAttachGnbWithConfiguration()
	},
}

func init() {
	rootCmd.AddCommand(gnbCmd)
}
