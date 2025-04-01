/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package tools

import (
	"my5G-RANTester/config"
	"net/netip"
)

func GenerateDefaultConf(controlIF netip.AddrPort, dataIF netip.AddrPort, amfs []*config.AMF) config.Config {
	return config.Config{
		GNodeB: config.GNodeB{
			ControlIF: config.IPv4Port{AddrPort: controlIF},
			DataIF:    config.IPv4Port{AddrPort: dataIF},
			PlmnList: config.PlmnList{
				Mcc:   "999",
				Mnc:   "70",
				Tac:   "000001",
				GnbId: "000008",
			},
			SliceSupportList: config.SliceSupportList{
				Sst: "01",
				Sd:  "000001",
			},
		},
		Ue: config.Ue{
			Msin:             "0000000120",
			Key:              "00112233445566778899AABBCCDDEEFF",
			Opc:              "00112233445566778899AABBCCDDEEFF",
			Amf:              "8000",
			Sqn:              "00000000",
			Dnn:              "internet",
			RoutingIndicator: "4567",
			Hplmn: config.Hplmn{
				Mcc: "999",
				Mnc: "70",
			},
			Snssai: config.Snssai{
				Sst: 01,
				Sd:  "000001",
			},
			Integrity: config.Integrity{
				Nia0: false,
				Nia1: true,
				Nia2: true,
			},
			Ciphering: config.Ciphering{
				Nea0: false,
				Nea1: true,
				Nea2: true,
			},
		},
		AMFs: amfs,
		Logs: config.Logs{
			Level: 4,
		},
	}
}
