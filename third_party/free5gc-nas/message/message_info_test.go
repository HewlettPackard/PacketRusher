package message

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/free5gc/nas/ie"
)

func TestBriefPDUSessEstReq(t *testing.T) {
	testCases := []struct {
		name string
		msg  *PDUSessEstReq
		want string
	}{
		{
			name: "PDU ID & PTI only",
			msg: &PDUSessEstReq{
				PDUSessId: 8,
				PTI:       9,
			},
			want: "PDUSessEstReq: [PduSessId:8,PTI:9]",
		},
		{
			name: "PDU ID, PTI & PDU Session Type",
			msg: &PDUSessEstReq{
				PDUSessId: 8,
				PTI:       9,
				PDUSessType: &ie.PDUSessType{
					Value: ie.PDUSessType_IPv4,
				},
			},
			want: "PDUSessEstReq: [PduSessId:8,PTI:9]",
		},
		{
			name: "Not printing all attributes..",
			msg: &PDUSessEstReq{
				PDUSessId: 5,
				PTI:       1,
				IntegrityProtectionMaxDataRate: &ie.IntegrityProtectionMaxDataRate{
					Uplink:   0xff,
					Downlink: 0xff,
				},
				PDUSessType: &ie.PDUSessType{
					Value: ie.PDUSessType_IPv4,
				},
				SSCMode: &ie.SSCMode{
					Mode: ie.SSCMODE1,
				},
				Capability5GSM: &ie.Capability5GSM{
					Rqos:    false,
					MH6PDU:  true,
					EPTS1:   false,
					ATSSSST: 5,
					TPMIC:   true,
				},
				AlwaysonPDUSessReq: &ie.AlwaysonPDUSessReq{
					APSR: true,
				},
				ExtendedProtCfgOpts: &ie.ExtendedProtCfgOpts{
					FromMs: &ie.ExtCfgOptFromMs{
						IPv4LinkMTUReq: true,
						DNSV4Req:       true,
					},
				},
				SuggestedIfId: &ie.PDUAddr{
					IPv6IfId: []byte{
						0xaa, 0xbb, 0xcc, 0xdd, 0x12, 0x34, 0x56, 0x78,
					},
					IPv4: []byte{0xde, 0xad, 0xbe, 0xef},
					SMFIPv6LLA: []byte{
						0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88,
						0x78, 0x56, 0x34, 0x12, 0x12, 0x34, 0x56, 0x78,
					},
				},
			},
			want: "PDUSessEstReq: [PduSessId:5,PTI:1]",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fmt.Println(tc.name)
			require.Equal(t, tc.want, tc.msg.String())
		})
	}
}

func TestBriefPDUSessEstAccept(t *testing.T) {
	testCases := []struct {
		name string
		msg  *PDUSessEstAccept
		want string
	}{
		{
			name: "PDU ID & PTI only",
			msg: &PDUSessEstAccept{
				PDUSessId: 8,
				PTI:       9,
			},
			want: "PDUSessEstAccept: [PduSessId:8,PTI:9]",
		},
		{
			name: "PDU ID, PTI & SessAMBR",
			msg: &PDUSessEstAccept{
				PDUSessId: 7,
				PTI:       8,
				SessAMBR: &ie.SessAMBR{
					UnitDownlink:  ie.Rate_4Kbps,
					ValueDownlink: 10,
					UnitUplink:    ie.Rate_64Gbps,
					ValueUplink:   499,
				},
			},
			want: "PDUSessEstAccept: [PduSessId:7,PTI:8,SessAMBR[UL:499*64G,DL:10*4K]]",
		},
		{
			name: "Not printing all attributes..",
			msg: &PDUSessEstAccept{
				PDUSessId: 10,
				PTI:       111,
				SelectedSSCMode: &ie.SSCMode{
					Mode: ie.SSCMODE1,
				},
				SelectedPDUSessType: &ie.PDUSessType{
					Value: ie.PDUSessType_IPv4,
				},
				AuthoQosRules: &ie.QosRules{
					Rules: []ie.QosRule{{
						RuleId:       1,
						OpCode:       ie.OpCode_CreateNewQosRule,
						IsDefaultDQR: true,
						Precedence:   255,
						Segregation:  false,
						QFI:          9,
						PktFilterList: []ie.PacketFilter{{
							Dir: ie.PFD_BiDir,
							Id:  1,
							Contents: ie.PacketFilterContents{
								MatchAll:   true,
								RemoteAddr: "any",
								LocalAddr:  "assigned",
							},
						}},
					}, {
						RuleId:        2,
						OpCode:        ie.OpCode_CreateNewQosRule,
						IsDefaultDQR:  false,
						Precedence:    23,
						Segregation:   false,
						QFI:           3,
						PktFilterList: []ie.PacketFilter{},
					}},
				},
				SessAMBR: &ie.SessAMBR{
					UnitDownlink:  ie.Rate_1Kbps,
					ValueDownlink: 1000,
					UnitUplink:    ie.Rate_1Kbps,
					ValueUplink:   1000,
				},
				AuthoQosFlowDescs: &ie.QosFlowDescs{
					Descs: []ie.QosFlowDesc{
						{
							QFI:    9,
							OpCode: ie.QFD_Create,
							EBit:   1, // parameters list is included
							FiveQI: uint8(ie.Qfd_5QI9),
						},
						{
							QFI:    3,
							OpCode: ie.QFD_Create,
							EBit:   1, // parameters list is included
							FiveQI: uint8(ie.Qfd_5QI3),
						},
					},
				},
				PDUAddr: &ie.PDUAddr{
					IPv4: []byte{0x0a, 0x3c, 0x00, 0x06},
				},
				SNSSAI: &ie.SNSSAI{
					SST: 0x01,
					SD:  "010203",
				},
				ExtendedProtCfgOpts: &ie.ExtendedProtCfgOpts{
					FromNw: &ie.ExtCfgOptFromNw{
						IPv4LinkMTU: 1358,
						DNSIPv4Addr: net.IP{0x08, 0x08, 0x08, 0x08},
					},
				},
				DNN: &ie.DNN{
					Value: "internet",
				},
			},
			want: "PDUSessEstAccept: [PduSessId:10,PTI:111," +
				"SessAMBR[UL:1000*1K,DL:1000*1K]," +
				"add QosRule[ID:1,DQF:true,prec:255,QFI:9,MatchAll]," +
				"add QosRule[ID:2,DQF:false,prec:23,QFI:3]," +
				"add QFD[QFI:9,5QI:9],add QFD[QFI:3,5QI:3]]",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fmt.Println(tc.name)
			require.Equal(t, tc.want, tc.msg.String())
		})
	}
}

func TestBriefPDUSessEstRej(t *testing.T) {
	testCases := []struct {
		name string
		msg  *PDUSessEstRej
		want string
	}{
		{
			name: "PDU ID & PTI only",
			msg: &PDUSessEstRej{
				PDUSessId: 8,
				PTI:       9,
			},
			want: "PDUSessEstRej: [PduSessId:8,PTI:9]",
		},
		{
			name: "PDU ID, PTI & SessAMBR",
			msg: &PDUSessEstRej{
				PDUSessId: 7,
				PTI:       8,
				Cause5GSM: &ie.Cause5GSM{
					Value: 0x1a,
				},
			},
			want: "PDUSessEstRej: [PduSessId:7,PTI:8,Cause:Insufficient resources]",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fmt.Println(tc.name)
			require.Equal(t, tc.want, tc.msg.String())
		})
	}
}

func TestBriefPDUSessModReq(t *testing.T) {
	testCases := []struct {
		name string
		msg  *PDUSessModReq
		want string
	}{
		{
			name: "PDU ID & PTI only",
			msg: &PDUSessModReq{
				PDUSessId: 8,
				PTI:       9,
			},
			want: "PDUSessModReq: [PduSessId:8,PTI:9]",
		},
		{
			name: "PDU ID, PTI & Cause5GSM",
			msg: &PDUSessModReq{
				PDUSessId: 8,
				PTI:       9,
				Cause5GSM: &ie.Cause5GSM{
					Value: 0x1a,
				},
			},
			want: "PDUSessModReq: [PduSessId:8,PTI:9,Cause:Insufficient resources]",
		},
		{
			name: "Not printing all attributes..",
			msg: &PDUSessModReq{
				PDUSessId: 5,
				PTI:       1,
				IntegrityProtectionMaxDataRate: &ie.IntegrityProtectionMaxDataRate{
					Uplink:   0xff,
					Downlink: 0xff,
				},
				Capability5GSM: &ie.Capability5GSM{
					Rqos:    false,
					MH6PDU:  true,
					EPTS1:   false,
					ATSSSST: 5,
					TPMIC:   true,
				},
				Cause5GSM: &ie.Cause5GSM{
					Value: 0x1c,
				},
				AlwaysonPDUSessReq: &ie.AlwaysonPDUSessReq{
					APSR: false,
				},
				ExtendedProtCfgOpts: &ie.ExtendedProtCfgOpts{
					FromMs: &ie.ExtCfgOptFromMs{
						IPv4LinkMTUReq: true,
						DNSV4Req:       true,
					},
				},
				ReqQosRules: &ie.QosRules{
					Rules: []ie.QosRule{{
						RuleId:       1,
						OpCode:       ie.OpCode_CreateNewQosRule,
						IsDefaultDQR: true,
						Precedence:   255,
						Segregation:  false,
						QFI:          9,
						PktFilterList: []ie.PacketFilter{{
							Dir: ie.PFD_BiDir,
							Id:  1,
							Contents: ie.PacketFilterContents{
								MatchAll:   true,
								RemoteAddr: "any",
								LocalAddr:  "assigned",
							},
						}},
					}, {
						RuleId:        2,
						OpCode:        ie.OpCode_CreateNewQosRule,
						IsDefaultDQR:  false,
						Precedence:    23,
						Segregation:   false,
						QFI:           3,
						PktFilterList: []ie.PacketFilter{},
					}},
				},
				ReqQosFlowDescs: &ie.QosFlowDescs{
					Descs: []ie.QosFlowDesc{
						{
							QFI:    9,
							OpCode: ie.QFD_Create,
							EBit:   1, // parameters list is included
							FiveQI: uint8(ie.Qfd_5QI9),
						},
						{
							QFI:    3,
							OpCode: ie.QFD_Create,
							EBit:   1, // parameters list is included
							FiveQI: uint8(ie.Qfd_5QI3),
						},
					},
				},
			},
			want: "PDUSessModReq: [PduSessId:5,PTI:1," +
				"add QosRule[ID:1,DQF:true,prec:255,QFI:9,MatchAll]," +
				"add QosRule[ID:2,DQF:false,prec:23,QFI:3]," +
				"add QFD[QFI:9,5QI:9],add QFD[QFI:3,5QI:3]," +
				"Cause:Unknown PDU session type]",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fmt.Println(tc.name)
			require.Equal(t, tc.want, tc.msg.String())
		})
	}
}

func TestBriefPDUSessModCmd(t *testing.T) {
	testCases := []struct {
		name string
		msg  *PDUSessModCmd
		want string
	}{
		{
			name: "PDU ID & PTI only",
			msg: &PDUSessModCmd{
				PDUSessId: 8,
				PTI:       9,
			},
			want: "PDUSessModCmd: [PduSessId:8,PTI:9]",
		},
		{
			name: "PDU ID, PTI & Cause5GSM",
			msg: &PDUSessModCmd{
				PDUSessId: 8,
				PTI:       9,
				Cause5GSM: &ie.Cause5GSM{
					Value: 0x1a,
				},
			},
			want: "PDUSessModCmd: [PduSessId:8,PTI:9,Cause:Insufficient resources]",
		},
		{
			name: "Empty QosRule + QFD",
			msg: &PDUSessModCmd{
				PDUSessId:     5,
				PTI:           1,
				AuthoQosRules: &ie.QosRules{},
				AuthoQosFlowDescs: &ie.QosFlowDescs{
					Descs: []ie.QosFlowDesc{
						{
							QFI:    9,
							OpCode: ie.QFD_Create,
							EBit:   1, // parameters list is included
							FiveQI: uint8(ie.Qfd_5QI9),
						},
						{
							QFI:    3,
							OpCode: ie.QFD_Create,
							EBit:   1, // parameters list is included
							FiveQI: uint8(ie.Qfd_5QI3),
						},
					},
				},
			},
			want: "PDUSessModCmd: [PduSessId:5,PTI:1," +
				"add QFD[QFI:9,5QI:9],add QFD[QFI:3,5QI:3]]",
		},
		{
			name: "Not printing all attributes..",
			msg: &PDUSessModCmd{
				PDUSessId: 5,
				PTI:       1,
				Cause5GSM: &ie.Cause5GSM{
					Value: 0x1c,
				},
				ExtendedProtCfgOpts: &ie.ExtendedProtCfgOpts{
					FromMs: &ie.ExtCfgOptFromMs{
						IPv4LinkMTUReq: true,
						DNSV4Req:       true,
					},
				},
				AuthoQosRules: &ie.QosRules{
					Rules: []ie.QosRule{{
						RuleId:       1,
						OpCode:       ie.OpCode_CreateNewQosRule,
						IsDefaultDQR: true,
						Precedence:   255,
						Segregation:  false,
						QFI:          9,
						PktFilterList: []ie.PacketFilter{{
							Dir: ie.PFD_BiDir,
							Id:  1,
							Contents: ie.PacketFilterContents{
								MatchAll:   true,
								RemoteAddr: "any",
								LocalAddr:  "assigned",
							},
						}},
					}, {
						RuleId:        2,
						OpCode:        ie.OpCode_CreateNewQosRule,
						IsDefaultDQR:  false,
						Precedence:    23,
						Segregation:   false,
						QFI:           3,
						PktFilterList: []ie.PacketFilter{},
					}},
				},
				AuthoQosFlowDescs: &ie.QosFlowDescs{
					Descs: []ie.QosFlowDesc{
						{
							QFI:    9,
							OpCode: ie.QFD_Create,
							EBit:   1, // parameters list is included
							FiveQI: uint8(ie.Qfd_5QI9),
						},
						{
							QFI:    3,
							OpCode: ie.QFD_Create,
							EBit:   1, // parameters list is included
							FiveQI: uint8(ie.Qfd_5QI3),
						},
					},
				},
			},
			want: "PDUSessModCmd: [PduSessId:5,PTI:1," +
				"add QosRule[ID:1,DQF:true,prec:255,QFI:9,MatchAll]," +
				"add QosRule[ID:2,DQF:false,prec:23,QFI:3]," +
				"add QFD[QFI:9,5QI:9],add QFD[QFI:3,5QI:3]," +
				"Cause:Unknown PDU session type]",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fmt.Println(tc.name)
			require.Equal(t, tc.want, tc.msg.String())
		})
	}
}

func TestBriefPDUSessModComplete(t *testing.T) {
	testCases := []struct {
		name string
		msg  *PDUSessModComplete
		want string
	}{
		{
			name: "PDU ID & PTI only",
			msg: &PDUSessModComplete{
				PDUSessId: 3,
				PTI:       2,
			},
			want: "PDUSessModComplete: [PduSessId:3,PTI:2]",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fmt.Println(tc.name)
			require.Equal(t, tc.want, tc.msg.String())
		})
	}
}

func TestBriefPDUSessModRej(t *testing.T) {
	testCases := []struct {
		name string
		msg  *PDUSessModRej
		want string
	}{
		{
			name: "PDU ID & PTI only",
			msg: &PDUSessModRej{
				PDUSessId: 8,
				PTI:       9,
			},
			want: "PDUSessModRej: [PduSessId:8,PTI:9]",
		},
		{
			name: "PDU ID, PTI & SessAMBR",
			msg: &PDUSessModRej{
				PDUSessId: 7,
				PTI:       8,
				Cause5GSM: &ie.Cause5GSM{
					Value: 0x1a,
				},
			},
			want: "PDUSessModRej: [PduSessId:7,PTI:8,Cause:Insufficient resources]",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fmt.Println(tc.name)
			require.Equal(t, tc.want, tc.msg.String())
		})
	}
}

func TestBriefPDUSessRelReq(t *testing.T) {
	testCases := []struct {
		name string
		msg  *PDUSessRelReq
		want string
	}{
		{
			name: "PDU ID & PTI only",
			msg: &PDUSessRelReq{
				PDUSessId: 8,
				PTI:       9,
			},
			want: "PDUSessRelReq: [PduSessId:8,PTI:9]",
		},
		{
			name: "PDU ID, PTI & SessAMBR",
			msg: &PDUSessRelReq{
				PDUSessId: 7,
				PTI:       8,
				Cause5GSM: &ie.Cause5GSM{
					Value: 0x1a,
				},
			},
			want: "PDUSessRelReq: [PduSessId:7,PTI:8,Cause:Insufficient resources]",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fmt.Println(tc.name)
			require.Equal(t, tc.want, tc.msg.String())
		})
	}
}

func TestBriefPDUSessRelCmd(t *testing.T) {
	testCases := []struct {
		name string
		msg  *PDUSessRelCmd
		want string
	}{
		{
			name: "PDU ID & PTI only",
			msg: &PDUSessRelCmd{
				PDUSessId: 8,
				PTI:       9,
			},
			want: "PDUSessRelCmd: [PduSessId:8,PTI:9]",
		},
		{
			name: "PDU ID, PTI & SessAMBR",
			msg: &PDUSessRelCmd{
				PDUSessId: 7,
				PTI:       8,
				Cause5GSM: &ie.Cause5GSM{
					Value: 0x1a,
				},
			},
			want: "PDUSessRelCmd: [PduSessId:7,PTI:8,Cause:Insufficient resources]",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fmt.Println(tc.name)
			require.Equal(t, tc.want, tc.msg.String())
		})
	}
}

func TestBriefPDUSessRelComplete(t *testing.T) {
	testCases := []struct {
		name string
		msg  *PDUSessRelComplete
		want string
	}{
		{
			name: "PDU ID & PTI only",
			msg: &PDUSessRelComplete{
				PDUSessId: 8,
				PTI:       9,
			},
			want: "PDUSessRelComplete: [PduSessId:8,PTI:9]",
		},
		{
			name: "PDU ID, PTI & SessAMBR",
			msg: &PDUSessRelComplete{
				PDUSessId: 7,
				PTI:       8,
				Cause5GSM: &ie.Cause5GSM{
					Value: 0x1a,
				},
			},
			want: "PDUSessRelComplete: [PduSessId:7,PTI:8,Cause:Insufficient resources]",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fmt.Println(tc.name)
			require.Equal(t, tc.want, tc.msg.String())
		})
	}
}

func TestBriefPDUSessRelRej(t *testing.T) {
	testCases := []struct {
		name string
		msg  *PDUSessRelRej
		want string
	}{
		{
			name: "PDU ID & PTI only",
			msg: &PDUSessRelRej{
				PDUSessId: 8,
				PTI:       9,
			},
			want: "PDUSessRelRej: [PduSessId:8,PTI:9]",
		},
		{
			name: "PDU ID, PTI & SessAMBR",
			msg: &PDUSessRelRej{
				PDUSessId: 7,
				PTI:       8,
				Cause5GSM: &ie.Cause5GSM{
					Value: 0x1a,
				},
			},
			want: "PDUSessRelRej: [PduSessId:7,PTI:8,Cause:Insufficient resources]",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fmt.Println(tc.name)
			require.Equal(t, tc.want, tc.msg.String())
		})
	}
}
