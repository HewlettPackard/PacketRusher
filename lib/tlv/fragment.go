package tlv

type fragments map[int][][]byte

func (f fragments) Add(tag int, buf []byte) {
	f[tag] = append(f[tag], buf)
}

func (f fragments) Get(tag int) ([][]byte, bool) {
	ret, t := f[tag]
	return ret, t
}
