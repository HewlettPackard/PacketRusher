package message

import (
	"fmt"
)

/* This file implements stringify functions for NAS messages.
   Since the go files in this directory are auto-generated,
   use a separated file to add the String() functions.
   Only fields of interest of MSGs of interest are implemented.
*/

func concat(s1, s2 string) string {
	if s1 != "" && s2 != "" {
		s1 += ","
	}
	return s1 + s2
}

func (m *PDUSessEstReq) String() string {
	str := fmt.Sprintf("%s: [PduSessId:%d,PTI:%d",
		m.MsgType(), m.PDUSessId, m.PTI)
	str += "]"
	return str
}

func (m *PDUSessEstAccept) String() string {
	str := fmt.Sprintf("%s: [PduSessId:%d,PTI:%d",
		m.MsgType(), m.PDUSessId, m.PTI)
	if ie := m.SessAMBR; ie != nil {
		str = concat(str, ie.String())
	}
	if ie := m.AuthoQosRules; ie != nil {
		str = concat(str, ie.String())
	}
	if ie := m.AuthoQosFlowDescs; ie != nil {
		str = concat(str, ie.String())
	}
	if ie := m.Cause5GSM; ie != nil {
		str = concat(str, fmt.Sprintf("Cause:%s", ie))
	}
	str += "]"
	return str
}

func (m PDUSessEstRej) String() string {
	str := fmt.Sprintf("%s: [PduSessId:%d,PTI:%d",
		m.MsgType(), m.PDUSessId, m.PTI)
	if ie := m.Cause5GSM; ie != nil {
		str = concat(str, fmt.Sprintf("Cause:%s", ie))
	}
	str += "]"
	return str
}

func (m *PDUSessModReq) String() string {
	str := fmt.Sprintf("%s: [PduSessId:%d,PTI:%d",
		m.MsgType(), m.PDUSessId, m.PTI)
	if ie := m.ReqQosRules; ie != nil {
		str = concat(str, ie.String())
	}
	if ie := m.ReqQosFlowDescs; ie != nil {
		str = concat(str, ie.String())
	}
	if ie := m.Cause5GSM; ie != nil {
		str = concat(str, fmt.Sprintf("Cause:%s", ie))
	}
	str += "]"
	return str
}

func (m *PDUSessModCmd) String() string {
	str := fmt.Sprintf("%s: [PduSessId:%d,PTI:%d",
		m.MsgType(), m.PDUSessId, m.PTI)
	if ie := m.SessAMBR; ie != nil {
		str = concat(str, ie.String())
	}
	if ie := m.AuthoQosRules; ie != nil {
		str = concat(str, ie.String())
	}
	if ie := m.AuthoQosFlowDescs; ie != nil {
		str = concat(str, ie.String())
	}
	if ie := m.Cause5GSM; ie != nil {
		str = concat(str, fmt.Sprintf("Cause:%s", ie))
	}
	str += "]"
	return str
}

func (m *PDUSessModComplete) String() string {
	str := fmt.Sprintf("%s: [PduSessId:%d,PTI:%d",
		m.MsgType(), m.PDUSessId, m.PTI)
	str += "]"
	return str
}

func (m *PDUSessModRej) String() string {
	str := fmt.Sprintf("%s: [PduSessId:%d,PTI:%d",
		m.MsgType(), m.PDUSessId, m.PTI)
	if ie := m.Cause5GSM; ie != nil {
		str = concat(str, fmt.Sprintf("Cause:%s", ie))
	}
	str += "]"
	return str
}

func (m *PDUSessRelReq) String() string {
	str := fmt.Sprintf("%s: [PduSessId:%d,PTI:%d",
		m.MsgType(), m.PDUSessId, m.PTI)
	if ie := m.Cause5GSM; ie != nil {
		str = concat(str, fmt.Sprintf("Cause:%s", ie))
	}
	str += "]"
	return str
}

func (m *PDUSessRelCmd) String() string {
	str := fmt.Sprintf("%s: [PduSessId:%d,PTI:%d",
		m.MsgType(), m.PDUSessId, m.PTI)
	if ie := m.Cause5GSM; ie != nil {
		str = concat(str, fmt.Sprintf("Cause:%s", ie))
	}
	str += "]"
	return str
}

func (m *PDUSessRelComplete) String() string {
	str := fmt.Sprintf("%s: [PduSessId:%d,PTI:%d",
		m.MsgType(), m.PDUSessId, m.PTI)
	if ie := m.Cause5GSM; ie != nil {
		str = concat(str, fmt.Sprintf("Cause:%s", ie))
	}
	str += "]"
	return str
}

func (m *PDUSessRelRej) String() string {
	str := fmt.Sprintf("%s: [PduSessId:%d,PTI:%d",
		m.MsgType(), m.PDUSessId, m.PTI)
	if ie := m.Cause5GSM; ie != nil {
		str = concat(str, fmt.Sprintf("Cause:%s", ie))
	}
	str += "]"
	return str
}
