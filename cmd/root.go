/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2024 Hewlett Packard Enterprise Development LP
 */
package cmd

import (
	"my5G-RANTester/config"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const version = "1.0.1"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "packetrusher",
	Short: "PacketRusher is a tool, based upon my5G-RANTester, dedicated to the performance testing and automatic validation of 5G Core Networks using simulated UE (user equipment) and gNodeB (5G base station).",
	Long:  `PacketRusher is a tool, based upon my5G-RANTester, dedicated to the performance testing and automatic validation of 5G Core Networks using simulated UE (user equipment) and gNodeB (5G base station).`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().String("config", "", "Configuration file path. (Default: ./config/config.yml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	cobra.OnInitialize(func() { log.Info("PacketRusher version " + version) })
}

func setConfig(path string) config.Config {
	var cfg config.Config
	if path == "" {
		cfg = config.LoadDefaultConfig()
	} else {
		cfg = config.Load(path)
	}
	return cfg
}
