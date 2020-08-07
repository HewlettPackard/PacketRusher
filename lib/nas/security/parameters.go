package security

// TS 33501 Annex A.8 Algorithm distinguisher For Knas_int Knas_enc
const (
	NNASEncAlg uint8 = 0x01
	NNASIntAlg uint8 = 0x02
	NRRCEncAlg uint8 = 0x03
	NRRCIntAlg uint8 = 0x04
	NUpEncAlg  uint8 = 0x05
	NUpIntAlg  uint8 = 0x06
)

// TS 33501 Annex D Algorithm identifier values For Knas_int
const (
	AlgIntegrity128NIA0 uint8 = 0x00 // NULL
	AlgIntegrity128NIA1 uint8 = 0x01 // 128-Snow3G
	AlgIntegrity128NIA2 uint8 = 0x02 // 128-AES
	AlgIntegrity128NIA3 uint8 = 0x03 // 128-ZUC
)

// TS 33501 Annex D Algorithm identifier values For Knas_enc
const (
	AlgCiphering128NEA0 uint8 = 0x00 // NULL
	AlgCiphering128NEA1 uint8 = 0x01 // 128-Snow3G
	AlgCiphering128NEA2 uint8 = 0x02 // 128-AES
	AlgCiphering128NEA3 uint8 = 0x03 // 128-ZUC
)

// 1bit
const (
	DirectionUplink   uint8 = 0x00
	DirectionDownlink uint8 = 0x01
)

// 5bits
const (
	OnlyOneBearer uint8 = 0x00
	Bearer3GPP    uint8 = 0x01
	BearerNon3GPP uint8 = 0x02
)

// TS 33501 Annex A.0 Access type distinguisher For Kgnb Kn3iwf
const (
	AccessType3GPP    uint8 = 0x01
	AccessTypeNon3GPP uint8 = 0x02
)
