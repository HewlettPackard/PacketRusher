package nasConvert

import (
	"my5G-RANTester/lib/nas/nasType"
)

// TS 24.501 9.11.3.35, TS 24.008 10.5.3.5a
func FullNetworkNameToNas(name string) (fullNetworkName nasType.FullNameForNetwork) {

	asciiArray := []byte(name)
	numOfSpareBits := 8 - ((7 * len(asciiArray)) % 8)

	var buf []uint8
	idx := uint8(7)
	for i, char := range asciiArray {
		if i == 0 {
			buf = append(buf, char)
		} else {
			buf[i-1] = (buf[i-1] & nasType.GetBitMask(idx+1, 0)) + uint8(char<<idx)
			buf = append(buf, char>>(8-idx))
			idx--
			// if idx overflow, it will round to max(uint8) == 255 == ^uint8(0)
			if idx == ^uint8(0) {
				idx = 7
			}
		}
	}

	fullNetworkName.SetLen(uint8(1 + len(buf)))
	fullNetworkName.SetCodingScheme(0)
	fullNetworkName.SetAddCI(0)
	fullNetworkName.SetExt(1)
	fullNetworkName.SetNumberOfSpareBitsInLastOctet(uint8(numOfSpareBits))
	fullNetworkName.SetTextString(buf)
	return
}

func ShortNetworkNameToNas(name string) (shortNetworkName nasType.ShortNameForNetwork) {

	asciiArray := []byte(name)
	numOfSpareBits := 8 - ((7 * len(asciiArray)) % 8)

	var buf []uint8
	idx := uint8(7)
	for i, char := range asciiArray {
		if i == 0 {
			buf = append(buf, char)
		} else {
			buf[i-1] = (buf[i-1] & nasType.GetBitMask(idx+1, 0)) + uint8(char<<idx)
			buf = append(buf, char>>(8-idx))
			idx--
			// if idx overflow, it will round to max(uint8) == 255 == ^uint8(0)
			if idx == ^uint8(0) {
				idx = 7
			}
		}
	}

	shortNetworkName.SetLen(uint8(1 + len(buf)))
	shortNetworkName.SetCodingScheme(0)
	shortNetworkName.SetAddCI(0)
	shortNetworkName.SetExt(1)
	shortNetworkName.SetNumberOfSpareBitsInLastOctet(uint8(numOfSpareBits))
	shortNetworkName.SetTextString(buf)
	return
}
