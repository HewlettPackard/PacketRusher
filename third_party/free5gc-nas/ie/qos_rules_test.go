package ie

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQosRulesUnmarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    []byte
		expected *QosRules
	}{
		{
			name: "Positive Case 1 ",
			input: []byte{
				0x01, 0x00, 0x06, 0x31,
				0x31, 0x01,
				0x01,
				0xff, 0x09,
			},
			expected: &QosRules{
				Rules: []QosRule{{
					RuleId:       1,
					OpCode:       OpCode_CreateNewQosRule,
					IsDefaultDQR: true,
					Precedence:   255,
					Segregation:  false,
					QFI:          9,
					PktFilterList: []PacketFilter{{
						Dir: PFD_BiDir,
						Id:  1,
						Contents: PacketFilterContents{
							MatchAll:   true,
							RemoteAddr: "any",
							LocalAddr:  "assigned",
						},
					}},
				}},
			},
		}, {
			name: "Delete Case - Positive",
			input: []byte{
				0x03, 0x00, 0x01, 0x40,
			},
			expected: &QosRules{
				Rules: []QosRule{{
					RuleId:        3,
					OpCode:        OpCode_DelExistingQosRule,
					IsDefaultDQR:  false,
					Precedence:    0,
					Segregation:   false,
					QFI:           0,
					PktFilterList: nil,
				}},
			},
		}, {
			name: "Delete + Create case",
			input: []byte{
				0x02, 0x00, 0x01, 0x40, 0x03, 0x00, 0x0e, 0x21, 0x31, 0x09, 0x10, 0x7f, 0x00,
				0x00, 0x01, 0xff, 0xff, 0xff, 0xff, 0xfc, 0x02,
			},
			expected: &QosRules{
				Rules: []QosRule{{
					RuleId:        2,
					OpCode:        OpCode_DelExistingQosRule,
					IsDefaultDQR:  false,
					Precedence:    0,
					Segregation:   false,
					QFI:           0,
					PktFilterList: nil,
				}, {
					RuleId:       3,
					OpCode:       OpCode_CreateNewQosRule,
					IsDefaultDQR: false,
					Precedence:   252,
					Segregation:  false,
					QFI:          2,
					PktFilterList: []PacketFilter{{
						Dir: PFD_BiDir,
						Id:  1,
						Contents: PacketFilterContents{
							RemoteAddr: "127.0.0.1/32",
							LocalAddr:  "assigned",
						},
					}},
				}},
			},
		}, {
			name: "Delete Case - Negative",
			input: []byte{
				0x03, 0x00, 0x02, 0x40, 0x00,
			},
		}, {
			name: "Create 1*QoSRule 1*PktFltr 1*Component",
			input: []byte{
				0x00, 0x00, 0x0E, 0x21,
				0x30, 0x09,
				0x10,                   // PFCType_Ipv4RemoteAddr
				0xAC, 0x10, 0x06, 0x01, // 172.16.6.1
				0xFF, 0xFF, 0xFF, 0xFF, // 255.255.255.255
				0xFF, 0x00,
			},
			expected: &QosRules{
				Rules: []QosRule{{
					RuleId:       0,
					OpCode:       OpCode_CreateNewQosRule,
					IsDefaultDQR: false,
					Precedence:   255,
					Segregation:  false,
					QFI:          0,
					PktFilterList: []PacketFilter{{
						Dir: PFD_BiDir,
						Id:  0,
						Contents: PacketFilterContents{
							RemoteAddr: "172.16.6.1/32",
							LocalAddr:  "assigned",
						},
					}},
				}},
			},
		}, {
			name: "Create 1*QoSRule 1*PktFltr 2*Components",
			input: []byte{
				// 1B QoS Rule, 2B Len, 3b operation(3) | 1b DQR(1) | 4b nr of pkt fltr
				0x01, 0x00, 0x13, 0x31,
				0x31, 0x0E,
				0x11,                   // PFCType_Ipv4LocalAddr
				0x0A, 0x0A, 0x00, 0x00, // 10.10.0.0
				0xFF, 0xFF, 0xFF, 0x00, // 255.255.255.0
				0x41,                   // Local port range
				0x03, 0xE8, 0x0B, 0xB8, // 0x03E8 ~ 0x0BB8
				0xFF, 0x09,
			},
			expected: &QosRules{
				Rules: []QosRule{{
					RuleId:       1,
					OpCode:       OpCode_CreateNewQosRule,
					IsDefaultDQR: true,
					Precedence:   255,
					Segregation:  false,
					QFI:          9,
					PktFilterList: []PacketFilter{{
						Dir: PFD_BiDir,
						Id:  1,
						Contents: PacketFilterContents{
							RemoteAddr:     "any",
							LocalAddr:      "10.10.0.0/24",
							LocalPortRange: "1000-3000",
						},
					}},
				}},
			},
		}, {
			name: "Create 1*QoSRule 2*PktFltr, 2*Component",
			input: []byte{
				// 1B QoS Rule, 2B Len, 3b operation(3) | 1b DQR(1) | 4b nr of pkt fltr
				0x01, 0x00, 0x21, 0x32,
				0x21, 0x0E,
				0x11,                   // PFCType_Ipv4LocalAddr
				0x0A, 0x0A, 0x00, 0x00, // 10.10.0.0
				0xFF, 0xFF, 0xFF, 0x00, // 255.255.255.0
				0x41,                   // Local port range
				0x03, 0xE8, 0x0B, 0xB8, // 0x03E8(1000) ~ 0x0BB8(3000)
				0x12, 0x0C,
				0x10,                   // PFCType_Ipv4RemoteAddr
				0x0A, 0x0A, 0x64, 0x00, // 10.10.100.0
				0xFF, 0xFF, 0xFF, 0x00, // 255.255.255.0
				0x50,       // PFCType_SingleRemotePort
				0x22, 0x65, // 0x2265 (8805)
				0xFF, 0x09,
			},
			expected: &QosRules{
				Rules: []QosRule{{
					RuleId:       1,
					OpCode:       OpCode_CreateNewQosRule,
					IsDefaultDQR: true,
					Precedence:   255,
					Segregation:  false,
					QFI:          9,
					PktFilterList: []PacketFilter{
						{
							Dir: PFD_Uplink,
							Id:  1,
							Contents: PacketFilterContents{
								RemoteAddr:     "any",
								LocalAddr:      "10.10.0.0/24",
								LocalPortRange: "1000-3000",
							},
						}, {
							Dir: PFD_Downlink,
							Id:  2,
							Contents: PacketFilterContents{
								RemoteAddr:      "10.10.100.0/24",
								LocalAddr:       "assigned",
								RemotePortRange: "8805",
							},
						},
					},
				}},
			},
		}, {
			name: "Create 2*QoSRule 2*PktFltr 2*Component",
			input: []byte{
				// 1B 1st QoS Rule, 2B Len, 3b operation(3) | 1b DQR(1) | 4b nr of pkt fltr
				0x01, 0x00, 0x21, 0x32,
				// bit65: dir, bit41:PktFltrId; len of PktFltrContent
				0x21, 0x0E,
				0x11,                   // PFCType_Ipv4LocalAddr
				0x0A, 0x0A, 0x00, 0x00, // 10.10.0.0
				0xFF, 0xFF, 0xFF, 0x00, // 255.255.255.0
				0x41,                   // Local port range
				0x03, 0xE8, 0x0B, 0xB8, // 0x03E8(1000) ~ 0x0BB8(3000)
				// bit65: dir, bit41:PktFltrId; len of PktFltrContent
				0x12, 0x0C,
				0x10,                   // PFCType_Ipv4RemoteAddr
				0x0A, 0x0A, 0x64, 0x00, // 10.10.100.0
				0xFF, 0xFF, 0xFF, 0x00, // 255.255.255.0
				0x50,       // PFCType_SingleRemotePort
				0x22, 0x65, // 0x2265 (8805)
				// QoS rule precedence, Segregation | QFI(6)
				0xFF, 0x09,

				// 1B 2nd QoS Rule, 2B Len, 3b operation(3) | 1b DQR(1) | 4b nr of pkt fltr
				0x02, 0x00, 0x2a, 0x23,
				// bit65: dir, bit41:PktFltrId; len of PktFltrContent
				0x21, 0x0E,
				0x11,                   // PFCType_Ipv4LocalAddr
				0x0A, 0x0A, 0x00, 0x00, // 10.10.0.0
				0xFF, 0xFF, 0xFF, 0x00, // 255.255.255.0
				0x41,                   // Local port range
				0x03, 0xE8, 0x0B, 0xB8, // 0x03E8 ~ 0x0BB8
				// bit65: dir, bit41:PktFltrId; len of PktFltrContent
				0x12, 0x0C,
				0x10,                   // PFCType_Ipv4RemoteAddr
				0x0A, 0x0A, 0x64, 0x00, // 10.10.100.0
				0xFF, 0xFF, 0xFF, 0x00, // 255.255.255.0
				0x50,       // PFCType_SingleRemotePort
				0x22, 0x65, // 0x2265 (8805)
				// bit65: dir, bit41:PktFltrId; len of PktFltrContent
				0x33, 0x07,
				0x82,                               // PFCType_SrcMacAddrType
				0x11, 0x3E, 0x4D, 0x02, 0x1B, 0xCD, // 11:3E:4d:02:1b:cd
				// QoS rule precedence, Segregation | QFI(6)
				0xFF, 0x05,
			},
			expected: &QosRules{
				Rules: []QosRule{
					{
						RuleId:       1,
						OpCode:       OpCode_CreateNewQosRule,
						IsDefaultDQR: true,
						Precedence:   255,
						Segregation:  false,
						QFI:          9,
						PktFilterList: []PacketFilter{
							{
								Dir: PFD_Uplink,
								Id:  1,
								Contents: PacketFilterContents{
									RemoteAddr:     "any",
									LocalAddr:      "10.10.0.0/24",
									LocalPortRange: "1000-3000",
								},
							}, {
								Dir: PFD_Downlink,
								Id:  2,
								Contents: PacketFilterContents{
									RemoteAddr:      "10.10.100.0/24",
									LocalAddr:       "assigned",
									RemotePortRange: "8805",
								},
							},
						},
					}, {
						RuleId:       2,
						OpCode:       OpCode_CreateNewQosRule,
						IsDefaultDQR: false,
						Precedence:   255,
						Segregation:  false,
						QFI:          5,
						PktFilterList: []PacketFilter{
							{
								Dir: PFD_Uplink,
								Id:  1,
								Contents: PacketFilterContents{
									RemoteAddr:     "any",
									LocalAddr:      "10.10.0.0/24",
									LocalPortRange: "1000-3000",
								},
							}, {
								Dir: PFD_Downlink,
								Id:  2,
								Contents: PacketFilterContents{
									RemoteAddr:      "10.10.100.0/24",
									LocalAddr:       "assigned",
									RemotePortRange: "8805",
								},
							}, {
								Dir: PFD_BiDir,
								Id:  3,
								Contents: PacketFilterContents{
									RemoteAddr: "any",
									LocalAddr:  "assigned",
									SrcMacAddr: "11:3e:4d:02:1b:cd",
								},
							},
						},
					},
				},
			},
		}, {
			name:  "Negative Case 1 - invalid length",
			input: []byte{0x01, 0x00, 0x04},
		}, {
			name: "Negative case 2 invalid length",
			input: []byte{
				// 1B 1st QoS Rule, 2B Len, 3b operation(3) | 1b DQR(1) | 4b nr of pkt fltr
				0x01, 0x00, 0x21, 0x32,
				// bit65: dir, bit41:PktFltrId; len of PktFltrContent
				0x21, 0x0d, // <--- invalid length here 0x0d -> 0x0E
				0x11, 0x0A, 0x0A, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0x00,
				0x41, 0x03, 0xE8, 0x0B, 0xB8,
				// bit65: dir, bit41:PktFltrId; len of PktFltrContent
				0x12, 0x0C,
				0x10, 0x0A, 0x0A, 0x64, 0x00, 0xFF, 0xFF, 0xFF, 0x00,
				0x50, 0x22, 0x65,
				// QoS rule precedence, Segregation | QFI(6)
				0xFF, 0x09,

				// 1B 2nd QoS Rule, 2B Len, 3b operation(3) | 1b DQR(1) | 4b nr of pkt fltr
				0x02, 0x00, 0x2a, 0x23,
				// bit65: dir, bit41:PktFltrId; len of PktFltrContent
				0x21, 0x0E,
				0x11, 0x0A, 0x0A, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0x00,
				0x41, 0x03, 0xE8, 0x0B, 0xB8,
				// bit65: dir, bit41:PktFltrId; len of PktFltrContent
				0x12, 0x0C,
				0x10, 0x0A, 0x0A, 0x64, 0x00, 0xFF, 0xFF, 0xFF, 0x00,
				0x50, 0x22, 0x65,
				// bit65: dir, bit41:PktFltrId; len of PktFltrContent
				0x33, 0x07,
				0x82, 0x11, 0x3E, 0x4D, 0x02, 0x1B, 0xCD,
				// QoS rule precedence, Segregation | QFI(6)
				0xFF, 0x05,
			},
		}, {
			name: "Negative case 2 invalid length",
			input: []byte{
				// 1B 1st QoS Rule, 2B Len, 3b operation(3) | 1b DQR(1) | 4b nr of pkt fltr
				0x01, 0x00, 0x21, 0x32,
				// bit65: dir, bit41:PktFltrId; len of PktFltrContent
				0x21, 0x0E,
				0x11, 0x0A, 0x0A, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0x00,
				0x41, 0x03, 0xE8, 0x0B, 0xB8,
				// bit65: dir, bit41:PktFltrId; len of PktFltrContent
				0x12, 0x0a, // <--- invalid length here 0x0a -> 0x0c
				0x10, 0x0A, 0x0A, 0x64, 0x00, 0xFF, 0xFF, 0xFF, 0x00,
				0x50, 0x22, 0x65,
				// QoS rule precedence, Segregation | QFI(6)
				0xFF, 0x09,

				// 1B 2nd QoS Rule, 2B Len, 3b operation(3) | 1b DQR(1) | 4b nr of pkt fltr
				0x02, 0x00, 0x2a, 0x23,
				// bit65: dir, bit41:PktFltrId; len of PktFltrContent
				0x21, 0x0E,
				0x11, 0x0A, 0x0A, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0x00,
				0x41, 0x03, 0xE8, 0x0B, 0xB8,
				// bit65: dir, bit41:PktFltrId; len of PktFltrContent
				0x12, 0x0C,
				0x10, 0x0A, 0x0A, 0x64, 0x00, 0xFF, 0xFF, 0xFF, 0x00,
				0x50, 0x22, 0x65,
				// bit65: dir, bit41:PktFltrId; len of PktFltrContent
				0x33, 0x07,
				0x82, 0x11, 0x3E, 0x4D, 0x02, 0x1B, 0xCD,
				// QoS rule precedence, Segregation | QFI(6)
				0xFF, 0x05,
			},
		}, {
			name: "Negative case 3 bad number of pkt filter",
			input: []byte{
				// 1B 1st QoS Rule, 2B Len, 3b operation(3) | 1b DQR(1) | 4b nr of pkt fltr
				0x01, 0x00, 0x21, 0x32,
				// bit65: dir, bit41:PktFltrId; len of PktFltrContent
				0x21, 0x0E,
				0x11, 0x0A, 0x0A, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0x00,
				0x41, 0x03, 0xE8, 0x0B, 0xB8,
				// bit65: dir, bit41:PktFltrId; len of PktFltrContent
				0x12, 0x0C,
				0x10, 0x0A, 0x0A, 0x64, 0x00, 0xFF, 0xFF, 0xFF, 0x00,
				0x50, 0x22, 0x65,
				// QoS rule precedence, Segregation | QFI(6)
				0xFF, 0x09,

				// 1B 2nd QoS Rule, 2B Len, 3b operation(3) | 1b DQR(1) | 4b nr of pkt fltr
				0x02, 0x00, 0x2a, 0x21, // <--- bad pkt filter 0x21  -> 0x23
				// bit65: dir, bit41:PktFltrId; len of PktFltrContent
				0x21, 0x0E,
				0x11, 0x0A, 0x0A, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0x00,
				0x41, 0x03, 0xE8, 0x0B, 0xB8,
				// bit65: dir, bit41:PktFltrId; len of PktFltrContent
				0x12, 0x0C,
				0x10, 0x0A, 0x0A, 0x64, 0x00, 0xFF, 0xFF, 0xFF, 0x00,
				0x50, 0x22, 0x65,
				// bit65: dir, bit41:PktFltrId; len of PktFltrContent
				0x33, 0x07,
				0x82, 0x11, 0x3E, 0x4D, 0x02, 0x1B, 0xCD,
				// QoS rule precedence, Segregation | QFI(6)
				0xFF, 0x05,
			},
		}, {
			name: "Negative case 3 bad length",
			input: []byte{
				// 1B 1st QoS Rule, 2B Len, 3b operation(3) | 1b DQR(1) | 4b nr of pkt fltr
				0x01, 0x00, 0x21, 0x32,
				// bit65: dir, bit41:PktFltrId; len of PktFltrContent
				0x21, 0x0E,
				0x11, 0x0A, 0x0A, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0x00,
				0x41, 0x03, 0xE8, 0x0B, 0xB8,
				// bit65: dir, bit41:PktFltrId; len of PktFltrContent
				0x12, 0x0C,
				0x10, 0x0A, 0x0A, 0x64, 0x00, 0xFF, 0xFF, 0xFF, 0x00,
				0x50, 0x22, 0x65,
				// QoS rule precedence, Segregation | QFI(6)
				0xFF, 0x09,

				// 1B 2nd QoS Rule, 2B Len, 3b operation(3) | 1b DQR(1) | 4b nr of pkt fltr
				0x02, 0x00, 0x2a, 0x23,
				// bit65: dir, bit41:PktFltrId; len of PktFltrContent
				0x21, 0x0E,
				0x11, 0x0A, 0x0A, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0x00,
				0x41, 0x03, 0xE8, 0x0B, 0xB8,
				// bit65: dir, bit41:PktFltrId; len of PktFltrContent
				0x12, 0x0C,
				0x10, 0x0A, 0x0A, 0x64, 0x00, 0xFF, 0xFF, 0xFF, 0x00,
				0x50, 0x22, 0x65,
				// bit65: dir, bit41:PktFltrId; len of PktFltrContent
				0x33, 0x09, // <-- bad length here, 0x09 -> 0x07
				0x82, 0x11, 0x3E, 0x4D, 0x02, 0x1B, 0xCD,
				// QoS rule precedence, Segregation | QFI(6)
				0xFF, 0x05,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(QosRules)
			err := ie.UnmarshalBinary(tc.input)
			if tc.expected != nil {
				require.NoError(t, err)
				require.Equal(t, tc.expected, ie)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestQosRulesMarshalBinary(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    *QosRules
		expected []byte
	}{
		{
			name: "Positive Case 1 ",
			expected: []byte{
				0x01, 0x00, 0x06, 0x31,
				0x31, 0x01,
				0x01,
				0xff, 0x09,
			},
			input: &QosRules{
				Rules: []QosRule{{
					RuleId:       1,
					OpCode:       OpCode_CreateNewQosRule,
					IsDefaultDQR: true,
					Precedence:   255,
					Segregation:  false,
					QFI:          9,
					PktFilterList: []PacketFilter{{
						Dir: PFD_BiDir,
						Id:  1,
						Contents: PacketFilterContents{
							MatchAll:   true,
							RemoteAddr: "any",
							LocalAddr:  "assigned",
						},
					}},
				}},
			},
		}, {
			name: "Delete Case",
			expected: []byte{
				0x03, 0x00, 0x01, 0x40,
			},
			input: &QosRules{
				Rules: []QosRule{{
					RuleId:        3,
					OpCode:        OpCode_DelExistingQosRule,
					IsDefaultDQR:  false,
					Precedence:    0,
					Segregation:   false,
					QFI:           0,
					PktFilterList: nil,
				}},
			},
		}, {
			name: "Delete Case with massive don't care fields",
			expected: []byte{
				0x03, 0x00, 0x01, 0x40,
			},
			input: &QosRules{
				Rules: []QosRule{{
					RuleId:       3,
					OpCode:       OpCode_DelExistingQosRule,
					IsDefaultDQR: false,
					Precedence:   230,
					Segregation:  false,
					QFI:          87,
					PktFilterList: []PacketFilter{{
						Dir: PFD_BiDir,
						Id:  0,
						Contents: PacketFilterContents{
							RemoteAddr: "172.16.6.1/32",
							LocalAddr:  "assigned",
						},
					}},
				}},
			},
		}, {
			name: "Create 1*QoSRule 1*PktFltr 1*Component",
			expected: []byte{
				0x00, 0x00, 0x0E, 0x21,
				0x30, 0x09,
				0x10,                   // PFCType_Ipv4RemoteAddr
				0xAC, 0x10, 0x06, 0x01, // 172.16.6.1
				0xFF, 0xFF, 0xFF, 0xFF, // 255.255.255.255
				0xFF, 0x00,
			},
			input: &QosRules{
				Rules: []QosRule{{
					RuleId:       0,
					OpCode:       OpCode_CreateNewQosRule,
					IsDefaultDQR: false,
					Precedence:   255,
					Segregation:  false,
					QFI:          0,
					PktFilterList: []PacketFilter{{
						Dir: PFD_BiDir,
						Id:  0,
						Contents: PacketFilterContents{
							RemoteAddr: "172.16.6.1/32",
							LocalAddr:  "assigned",
						},
					}},
				}},
			},
		}, {
			name: "Create 1*QoSRule 1*PktFltr 2*Components",
			expected: []byte{
				// 1B QoS Rule, 2B Len, 3b operation(3) | 1b DQR(1) | 4b nr of pkt fltr
				0x01, 0x00, 0x13, 0x31,
				0x31, 0x0E,
				0x11,                   // PFCType_Ipv4LocalAddr
				0x0A, 0x0A, 0x00, 0x00, // 10.10.0.0
				0xFF, 0xFF, 0xFF, 0x00, // 255.255.255.0
				0x41,                   // Local port range
				0x03, 0xE8, 0x0B, 0xB8, // 0x03E8 ~ 0x0BB8
				0xFF, 0x09,
			},
			input: &QosRules{
				Rules: []QosRule{{
					RuleId:       1,
					OpCode:       OpCode_CreateNewQosRule,
					IsDefaultDQR: true,
					Precedence:   255,
					Segregation:  false,
					QFI:          9,
					PktFilterList: []PacketFilter{{
						Dir: PFD_BiDir,
						Id:  1,
						Contents: PacketFilterContents{
							RemoteAddr:     "any",
							LocalAddr:      "10.10.0.0/24",
							LocalPortRange: "1000-3000",
						},
					}},
				}},
			},
		}, {
			name: "Create 1*QoSRule 2*PktFltr, 2*Component",
			expected: []byte{
				// 1B QoS Rule, 2B Len, 3b operation(3) | 1b DQR(1) | 4b nr of pkt fltr
				0x01, 0x00, 0x21, 0x32,
				0x21, 0x0E,
				0x11,                   // PFCType_Ipv4LocalAddr
				0x0A, 0x0A, 0x00, 0x00, // 10.10.0.0
				0xFF, 0xFF, 0xFF, 0x00, // 255.255.255.0
				0x41,                   // Local port range
				0x03, 0xE8, 0x0B, 0xB8, // 0x03E8(1000) ~ 0x0BB8(3000)
				0x12, 0x0C,
				0x10,                   // PFCType_Ipv4RemoteAddr
				0x0A, 0x0A, 0x64, 0x00, // 10.10.100.0
				0xFF, 0xFF, 0xFF, 0x00, // 255.255.255.0
				0x50,       // PFCType_SingleRemotePort
				0x22, 0x65, // 0x2265 (8805)
				0xFF, 0x09,
			},
			input: &QosRules{
				Rules: []QosRule{{
					RuleId:       1,
					OpCode:       OpCode_CreateNewQosRule,
					IsDefaultDQR: true,
					Precedence:   255,
					Segregation:  false,
					QFI:          9,
					PktFilterList: []PacketFilter{
						{
							Dir: PFD_Uplink,
							Id:  1,
							Contents: PacketFilterContents{
								RemoteAddr:     "any",
								LocalAddr:      "10.10.0.0/24",
								LocalPortRange: "1000-3000",
							},
						}, {
							Dir: PFD_Downlink,
							Id:  2,
							Contents: PacketFilterContents{
								RemoteAddr:      "10.10.100.0/24",
								LocalAddr:       "assigned",
								RemotePortRange: "8805",
							},
						},
					},
				}},
			},
		}, {
			name: "Create 2*QoSRule 2*PktFltr 2*Component",
			expected: []byte{
				// 1B 1st QoS Rule, 2B Len, 3b operation(3) | 1b DQR(1) | 4b nr of pkt fltr
				0x01, 0x00, 0x21, 0x32,
				// bit65: dir, bit41:PktFltrId; len of PktFltrContent
				0x21, 0x0E,
				0x11,                   // PFCType_Ipv4LocalAddr
				0x0A, 0x0A, 0x00, 0x00, // 10.10.0.0
				0xFF, 0xFF, 0xFF, 0x00, // 255.255.255.0
				0x41,                   // Local port range
				0x03, 0xE8, 0x0B, 0xB8, // 0x03E8(1000) ~ 0x0BB8(3000)
				// bit65: dir, bit41:PktFltrId; len of PktFltrContent
				0x12, 0x0C,
				0x10,                   // PFCType_Ipv4RemoteAddr
				0x0A, 0x0A, 0x64, 0x00, // 10.10.100.0
				0xFF, 0xFF, 0xFF, 0x00, // 255.255.255.0
				0x50,       // PFCType_SingleRemotePort
				0x22, 0x65, // 0x2265 (8805)
				// QoS rule precedence, Segregation | QFI(6)
				0xFF, 0x09,

				// 1B 2nd QoS Rule, 2B Len, 3b operation(3) | 1b DQR(1) | 4b nr of pkt fltr
				0x02, 0x00, 0x2a, 0x23,
				// bit65: dir, bit41:PktFltrId; len of PktFltrContent
				0x21, 0x0E,
				0x11,                   // PFCType_Ipv4LocalAddr
				0x0A, 0x0A, 0x00, 0x00, // 10.10.0.0
				0xFF, 0xFF, 0xFF, 0x00, // 255.255.255.0
				0x41,                   // Local port range
				0x03, 0xE8, 0x0B, 0xB8, // 0x03E8 ~ 0x0BB8
				// bit65: dir, bit41:PktFltrId; len of PktFltrContent
				0x12, 0x0C,
				0x10,                   // PFCType_Ipv4RemoteAddr
				0x0A, 0x0A, 0x64, 0x00, // 10.10.100.0
				0xFF, 0xFF, 0xFF, 0x00, // 255.255.255.0
				0x50,       // PFCType_SingleRemotePort
				0x22, 0x65, // 0x2265 (8805)
				// bit65: dir, bit41:PktFltrId; len of PktFltrContent
				0x33, 0x07,
				0x82,                               // PFCType_SrcMacAddrType
				0x11, 0x3E, 0x4D, 0x02, 0x1B, 0xCD, // 11:3E:4d:02:1b:cd
				// QoS rule precedence, Segregation | QFI(6)
				0xFF, 0x05,
			},
			input: &QosRules{
				Rules: []QosRule{
					{
						RuleId:       1,
						OpCode:       OpCode_CreateNewQosRule,
						IsDefaultDQR: true,
						Precedence:   255,
						Segregation:  false,
						QFI:          9,
						PktFilterList: []PacketFilter{
							{
								Dir: PFD_Uplink,
								Id:  1,
								Contents: PacketFilterContents{
									RemoteAddr:     "any",
									LocalAddr:      "10.10.0.0/24",
									LocalPortRange: "1000-3000",
								},
							}, {
								Dir: PFD_Downlink,
								Id:  2,
								Contents: PacketFilterContents{
									RemoteAddr:      "10.10.100.0/24",
									LocalAddr:       "assigned",
									RemotePortRange: "8805",
								},
							},
						},
					}, {
						RuleId:       2,
						OpCode:       OpCode_CreateNewQosRule,
						IsDefaultDQR: false,
						Precedence:   255,
						Segregation:  false,
						QFI:          5,
						PktFilterList: []PacketFilter{
							{
								Dir: PFD_Uplink,
								Id:  1,
								Contents: PacketFilterContents{
									RemoteAddr:     "any",
									LocalAddr:      "10.10.0.0/24",
									LocalPortRange: "1000-3000",
								},
							}, {
								Dir: PFD_Downlink,
								Id:  2,
								Contents: PacketFilterContents{
									RemoteAddr:      "10.10.100.0/24",
									LocalAddr:       "assigned",
									RemotePortRange: "8805",
								},
							}, {
								Dir: PFD_BiDir,
								Id:  3,
								Contents: PacketFilterContents{
									RemoteAddr: "any",
									LocalAddr:  "assigned",
									SrcMacAddr: "11:3e:4d:02:1b:cd",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := tc.input.MarshalBinary()
			if tc.expected != nil {
				require.NoError(t, err)
				require.Equal(t, tc.expected, b)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestPacketFilterContentsMarshalUnmarshal(t *testing.T) {
	testCases := []struct {
		name        string
		marshaled   []byte
		unmarshaled *PacketFilterContents
	}{
		{
			name: "match all",
			marshaled: []byte{
				0x01, // PFCType_MatchAllType
			},
			unmarshaled: &PacketFilterContents{
				MatchAll:   true,
				RemoteAddr: "any",
				LocalAddr:  "assigned",
			},
		},
		{
			name: "Src Mac Addr + Remote Mac Addr",
			marshaled: []byte{
				0x82, // PFCType_SrcMacAddr
				0x11, 0x3E, 0x4D, 0x02, 0x1B, 0xCD,
				0x81, // PFCType_DstMacAddr
				0x00, 0x40, 0x96, 0x57, 0xb4, 0xac,
			},
			unmarshaled: &PacketFilterContents{
				DstMacAddr: "00:40:96:57:b4:ac",
				// LocalAddr:  "assigned",
				SrcMacAddr: "11:3e:4d:02:1b:cd",
				RemoteAddr: "any",
				LocalAddr:  "assigned",
			},
		},
		{
			name: "Src Mac Addr Range + Remote Mac Addr Range",
			marshaled: []byte{
				0x89,                               // PFCType_SrcMACAddrRange
				0x11, 0x22, 0x33, 0xAA, 0xBB, 0xCC, //
				0x11, 0x3E, 0x4D, 0x02, 0x1B, 0xCD, //
				0x88, // PFCType_DstMACAddrRange
				0x00, 0x40, 0x96, 0x57, 0xb4, 0xac,
				0x11, 0x3E, 0x4D, 0x02, 0x1B, 0xCD, //
			},
			unmarshaled: &PacketFilterContents{
				DstMacAddrRange: [2]string{"00:40:96:57:b4:ac", "11:3e:4d:02:1b:cd"},
				SrcMacAddrRange: [2]string{"11:22:33:aa:bb:cc", "11:3e:4d:02:1b:cd"},
				RemoteAddr:      "any",
				LocalAddr:       "assigned",
			},
		},
		{
			name: "Single Port",
			marshaled: []byte{
				0x50,       // PFCType_SingleRemotePort
				0x4D, 0x02, //
				0x40,       // PFCType_SingleLocalPort
				0x11, 0x3E, //
			},
			unmarshaled: &PacketFilterContents{
				LocalPortRange:  "4414",
				RemotePortRange: "19714",
				LocalAddr:       "assigned",
				RemoteAddr:      "any",
			},
		},
		{
			name: "IPv4 + Port Range",
			marshaled: []byte{
				0x10,                   // PFCType_Ipv4RemoteAddr
				0x0a, 0x00, 0x00, 0x00, //
				0xff, 0xff, 0xff, 0x00, //
				0x11,                   // PFCType_Ipv4LocalAddr
				0xc0, 0xa8, 0x00, 0x00, //
				0xff, 0xff, 0xff, 0x80, //
				0x51,                   // PFCType_RemotePortRange
				0x0D, 0x02, 0x11, 0x3E, //
				0x41,                   // PFCType_LocalPortRange
				0x11, 0x3E, 0x4D, 0x02, //
			},
			unmarshaled: &PacketFilterContents{
				RemotePortRange: "3330-4414",
				LocalPortRange:  "4414-19714",
				RemoteAddr:      "10.0.0.0/24",
				LocalAddr:       "192.168.0.0/25",
			},
		},
		{
			name: "IPv6",
			marshaled: []byte{
				0x21,                                           // PFCType_Ipv6RemoteAddrPrefixLen
				0x20, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, //
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, //
				0x40,                                           // prefix length
				0x23,                                           // PFCType_Ipv6LocalAddrPrefixLen
				0x20, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, //
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, //
				0x40, // prefix length
			},
			unmarshaled: &PacketFilterContents{
				LocalAddr:  "2001::/64",
				RemoteAddr: "2001::/64",
			},
		},
		{
			name: "Flow Label + TOS + SPI",
			marshaled: []byte{
				0x60,                   // PFCType_SecurityParameterIndex
				0x01, 0x02, 0x03, 0x04, // IPSec Security parameter index
				0x70,       // PFCType_TypesOfService
				0x01, 0x02, // TOS/traffic class + TOS/traffic class mask
				0x80,             // PFCType_FlowLabel
				0x01, 0x02, 0x03, // bit 8-5 in 1st octet spare, remaining 20 bits the IPv6 flow label
			},
			unmarshaled: &PacketFilterContents{
				FlowLabel:       "10203",
				SPI:             "1020304",
				TosTrafficClass: "0102",
				LocalAddr:       "assigned",
				RemoteAddr:      "any",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ie := new(PacketFilterContents)
			err := ie.UnmarshalBinary(tc.marshaled)
			if tc.unmarshaled != nil {
				require.NoError(t, err)
				require.Equal(t, tc.unmarshaled, ie)
			} else {
				require.Error(t, err)
			}

			b, err := tc.unmarshaled.MarshalBinary()
			if err != nil {
				require.Error(t, err)
			} else {
				require.Equal(t, tc.marshaled, b)
			}
		})
	}
}

func TestBriefNasQosRules(t *testing.T) {
	testCases := []struct {
		name  string
		nasQR *QosRules
		want  string
	}{
		{
			name:  "case 1",
			nasQR: &QosRules{},
			want:  "",
		},
		{
			name: "case 2",
			nasQR: &QosRules{
				Rules: []QosRule{
					{},
				},
			},
			want: "",
		},
		{
			name: "case 3",
			nasQR: &QosRules{
				Rules: []QosRule{
					{},
					{
						RuleId:       1,
						IsDefaultDQR: true,
						Precedence:   1,
						QFI:          1,
						OpCode:       OpCode_CreateNewQosRule,
						PktFilterList: []PacketFilter{
							{
								Contents: PacketFilterContents{
									MatchAll:        true,
									LocalAddr:       "",
									LocalPortRange:  "",
									RemoteAddr:      "",
									RemotePortRange: "",
								},
							},
						},
					},
					{},
				},
			},
			want: "add QosRule[ID:1,DQF:true,prec:1,QFI:1,MatchAll]",
		},
		{
			name: "case 4",
			nasQR: &QosRules{
				Rules: []QosRule{
					{
						RuleId:        1,
						IsDefaultDQR:  true,
						Precedence:    1,
						QFI:           1,
						OpCode:        OpCode_CreateNewQosRule,
						PktFilterList: []PacketFilter{},
					},
				},
			},
			want: "add QosRule[ID:1,DQF:true,prec:1,QFI:1]",
		},
		{
			name: "case 5",
			nasQR: &QosRules{
				Rules: []QosRule{
					{
						RuleId:       1,
						IsDefaultDQR: true,
						Precedence:   1,
						QFI:          1,
						OpCode:       OpCode_CreateNewQosRule,
						PktFilterList: []PacketFilter{
							{
								Contents: PacketFilterContents{
									MatchAll:        true,
									LocalAddr:       "",
									LocalPortRange:  "",
									RemoteAddr:      "",
									RemotePortRange: "",
								},
							},
						},
					},
					{
						RuleId:       2,
						IsDefaultDQR: false,
						Precedence:   2,
						QFI:          2,
						OpCode:       OpCode_ModifyReplaceAllPktFilters,
						PktFilterList: []PacketFilter{
							{
								Contents: PacketFilterContents{
									MatchAll:        false,
									LocalAddr:       "192.168.0.1",
									LocalPortRange:  "1024-2048",
									RemoteAddr:      "10.0.0.1",
									RemotePortRange: "",
								},
							},
						},
					},
					{
						// In case of Del QosRule, Precedence should be 0,
						// but printing it out can detect bad values.
						RuleId:       3,
						IsDefaultDQR: false,
						Precedence:   254,
						QFI:          3,
						OpCode:       OpCode_DelExistingQosRule,
					},
				},
			},
			want: "add QosRule[ID:1,DQF:true,prec:1,QFI:1,MatchAll]," +
				"mod QosRule[ID:2,DQF:false,prec:2,QFI:2,10.0.0.1-->192.168.0.1:1024-2048]," +
				"del QosRule[ID:3,DQF:false,prec:254,QFI:3]",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fmt.Println(tc.name)
			require.Equal(t, tc.want, tc.nasQR.String())
		})
	}
}
