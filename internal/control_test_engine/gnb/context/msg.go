package context

type UEMessage struct {
	GnbIp string
	UpfIp string
	OTeid string
	ITeid string
}

func (ue *UEMessage) NewUeMessage(GnbIp string, UpfIp string, OTeid string, ITeid string) {
	ue.GnbIp = GnbIp
	ue.OTeid = OTeid
	ue.UpfIp = UpfIp
	ue.ITeid = ITeid
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