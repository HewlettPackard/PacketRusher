package context

type UEMessage struct {
	GNBPduSessions [16]*GnbPDUSession
	GnbIp string
	UpfIp string
	GNBRx chan UEMessage
	GNBTx chan UEMessage
	IsNas bool
	Nas   []byte
	ConnectionClosed bool
	AmfId int64
	Msin string
}
