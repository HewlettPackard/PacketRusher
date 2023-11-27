package tools

import (
	"my5G-RANTester/config"
)

func GenerateDefaultConf(controlIF config.ControlIF, dataIF config.DataIF, amf config.AMF) config.Config {
	return config.Config{
		GNodeB: config.GNodeB{
			ControlIF: controlIF,
			DataIF:    dataIF,
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
				Nia1: false,
				Nia2: true,
			},
			Ciphering: config.Ciphering{
				Nea0: false,
				Nea1: false,
				Nea2: true,
			},
		},
		AMF: amf,
		Logs: config.Logs{
			Level: 4,
		},
	}
}
