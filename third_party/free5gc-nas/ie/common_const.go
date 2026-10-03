package ie

type BitRateType uint8

const (
	Rate_1Kbps   BitRateType = 0x01
	Rate_4Kbps   BitRateType = 0x02
	Rate_16Kbps  BitRateType = 0x03
	Rate_64Kbps  BitRateType = 0x04
	Rate_256Kbps BitRateType = 0x05
	Rate_1Mbps   BitRateType = 0x06
	Rate_4Mbps   BitRateType = 0x07
	Rate_16Mbps  BitRateType = 0x08
	Rate_64Mbps  BitRateType = 0x09
	Rate_256Mbps BitRateType = 0x0a
	Rate_1Gbps   BitRateType = 0x0b
	Rate_4Gbps   BitRateType = 0x0c
	Rate_16Gbps  BitRateType = 0x0d
	Rate_64Gbps  BitRateType = 0x0e
	Rate_256Gbps BitRateType = 0x0f
	Rate_1Tbps   BitRateType = 0x10
	Rate_4Tbps   BitRateType = 0x11
	Rate_16Tbps  BitRateType = 0x12
	Rate_64Tbps  BitRateType = 0x13
	Rate_256Tbps BitRateType = 0x14
	Rate_1Pbps   BitRateType = 0x15
	Rate_4Pbps   BitRateType = 0x16
	Rate_16Pbps  BitRateType = 0x17
	Rate_64Pbps  BitRateType = 0x18
	Rate_256Pbps BitRateType = 0x19
)

var BitRateTypeStrings = map[BitRateType]string{
	Rate_1Kbps:   "1K",
	Rate_4Kbps:   "4K",
	Rate_16Kbps:  "16K",
	Rate_64Kbps:  "64K",
	Rate_256Kbps: "256K",
	Rate_1Mbps:   "1M",
	Rate_4Mbps:   "4M",
	Rate_16Mbps:  "16M",
	Rate_64Mbps:  "64M",
	Rate_256Mbps: "256M",
	Rate_1Gbps:   "1G",
	Rate_4Gbps:   "4G",
	Rate_16Gbps:  "16G",
	Rate_64Gbps:  "64G",
	Rate_256Gbps: "256G",
	Rate_1Tbps:   "1T",
	Rate_4Tbps:   "4T",
	Rate_16Tbps:  "16T",
	Rate_64Tbps:  "64T",
	Rate_256Tbps: "256T",
	Rate_1Pbps:   "1P",
	Rate_4Pbps:   "4P",
	Rate_16Pbps:  "16P",
	Rate_64Pbps:  "64P",
	Rate_256Pbps: "256P",
}

func BitRateStr(u BitRateType) string {
	if s, ok := BitRateTypeStrings[u]; ok {
		return s
	}
	return "Unknown"
}
