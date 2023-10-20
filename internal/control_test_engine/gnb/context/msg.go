package context

type UEMessage struct {
	PDUSessionId int64
	GnbIp string
	UpfIp string
	OTeid string
	ITeid string
	GNBRx chan UEMessage
	GNBTx chan UEMessage
	IsNas bool
	Nas   []byte
	ConnectionClosed bool
	AmfId int64
	Msin string
}

func (ue *UEMessage) NewUeMessage(PDUSessionId int64, GnbIp string, UpfIp string, OTeid string, ITeid string) {
	ue.PDUSessionId = PDUSessionId
	ue.GnbIp = GnbIp
	ue.OTeid = OTeid
	ue.UpfIp = UpfIp
	ue.ITeid = ITeid
}

func (ue *UEMessage) GetPDUSessionId() int64 {
	return ue.PDUSessionId
}

func (ue *UEMessage) GetGnbIp() string {
	return ue.GnbIp
}

func (ue *UEMessage) GetUpfIp() string {
	return ue.UpfIp
}

func (ue *UEMessage) GetOTeid() string {
	return ue.OTeid
}

func (ue *UEMessage) GetITeid() string {
	return ue.ITeid
}