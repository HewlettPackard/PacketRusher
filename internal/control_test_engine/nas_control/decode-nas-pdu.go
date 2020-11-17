package nas_control

import (
	"my5G-RANTester/lib/nas"
	"my5G-RANTester/lib/ngap/ngapType"
)

func GetNasPduFromDownlink(msg *ngapType.DownlinkNASTransport) (m *nas.Message) {
	for _, ie := range msg.ProtocolIEs.List {
		if ie.Id.Value == ngapType.ProtocolIEIDNASPDU {
			pkg := []byte(ie.Value.NASPDU.Value)
			m = new(nas.Message)
			err := m.PlainNasDecode(&pkg)
			if err != nil {
				return nil
			}
			return
		}
	}
	return nil
}

func GetNasPduFromPduAccept(dlNas *nas.Message) (m *nas.Message) {

	// get payload container from DL NAS.
	payload := dlNas.DLNASTransport.GetPayloadContainerContents()
	m = new(nas.Message)
	err := m.PlainNasDecode(&payload)
	if err != nil {
		return nil
	}
	return
}

func GetNasPduFromDlNas(msg *ngapType.PDUSessionResourceSetupRequest) (m *nas.Message) {
	for _, ie := range msg.ProtocolIEs.List {
		if ie.Id.Value == ngapType.ProtocolIEIDPDUSessionResourceSetupListSUReq {
			pDUSessionResourceSetupList := ie.Value.PDUSessionResourceSetupListSUReq
			for _, item := range pDUSessionResourceSetupList.List {
				// get PDUSessionNas-PDU
				payload := []byte(item.PDUSessionNASPDU.Value)
				// remove security header.
				payload = payload[7:]
				m := new(nas.Message)
				err := m.PlainNasDecode(&payload)
				if err != nil {
					return nil
				}
				return m
			}
		}
	}
	return nil
}
