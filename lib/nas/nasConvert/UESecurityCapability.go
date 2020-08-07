package nasConvert

func UESecurityCapabilityToByteArray(buf []uint8) (nea, nia, eea, eia [2]byte) {
	if len(buf) < 2 {
		return
	}
	nea[0] = buf[0] << 1
	nia[0] = buf[1] << 1
	if len(buf) > 2 {
		eea[0] = buf[2] << 1
		eia[0] = buf[3] << 1
	}
	return
}
