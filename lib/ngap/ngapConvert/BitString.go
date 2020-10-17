package ngapConvert

import (
	"encoding/hex"
	"my5G-RANTester/lib/aper"
)

func BitStringToHex(bitString *aper.BitString) (hexString string) {
	hexString = hex.EncodeToString(bitString.Bytes)
	hexLen := (bitString.BitLength + 3) / 4
	hexString = hexString[:hexLen]
	return
}

func HexToBitString(hexString string, bitLength int) (bitString aper.BitString) {
	hexLen := len(hexString)
	if hexLen != (bitLength+3)/4 {
		// logger.NgapLog.Warningln("hexLen[", hexLen, "] doesn't match bitLength[", bitLength, "]")
		return
	}
	if hexLen%2 == 1 {
		hexString += "0"
	}
	bitString.Bytes, _ = hex.DecodeString(hexString)
	bitString.BitLength = uint64(bitLength)
	mask := byte(0xff)
	mask = mask << uint(8-bitLength%8)
	if mask != 0 {
		bitString.Bytes[len(bitString.Bytes)-1] &= mask
	}
	return
}

func ByteToBitString(byteArray []byte, bitLength int) (bitString aper.BitString) {
	byteLen := (bitLength + 7) / 8
	if byteLen > len(byteArray) {
		// logger.NgapLog.Warningln("bitLength[", bitLength, "] is beyond byteArray size[", len(byteArray), "]")
		return
	}
	bitString.Bytes = byteArray
	bitString.BitLength = uint64(bitLength)
	return
}
