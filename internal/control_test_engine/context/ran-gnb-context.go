package context

import (
	"encoding/hex"
	"fmt"
)

type RanGnbContext struct {
	mcc   string
	mnc   string
	tac   string
	gnbId string
	// added other information about gnb
	slice struct {
		st  string
		sst string
	}
}

func (gnb *RanGnbContext) NewRanGnbContext(gnbId, mcc, mnc, tac, st, sst string) {
	gnb.mcc = mcc
	gnb.mnc = mnc
	gnb.tac = tac
	gnb.gnbId = gnbId
	gnb.slice.st = st
	gnb.slice.sst = sst
}

func reverse(s string) string {
	// reverse string.
	var aux string
	for _, valor := range s {
		aux = string(valor) + aux
	}
	return aux
}

func (gnb *RanGnbContext) getGnbId() string {
	return gnb.gnbId
}

func (gnb *RanGnbContext) GetGnbIdInBytes() []byte {
	// changed for bytes.
	resu, err := hex.DecodeString(gnb.gnbId)
	if err != nil {
		fmt.Println(err)
	}
	return resu
}

func (gnb *RanGnbContext) getTac() string {
	return gnb.tac
}

func (gnb *RanGnbContext) getTacInBytes() []byte {
	// changed for bytes.
	resu, err := hex.DecodeString(gnb.tac)
	if err != nil {
		fmt.Println(err)
	}
	return resu
}

func (gnb *RanGnbContext) getSlice() (string, string) {
	return gnb.slice.st, gnb.slice.sst
}

func (gnb *RanGnbContext) getSliceInBytes() ([]byte, []byte) {
	stBytes, err := hex.DecodeString(gnb.slice.st)
	if err != nil {
		fmt.Println(err)
	}

	sstBytes, err := hex.DecodeString(gnb.slice.sst)
	if err != nil {
		fmt.Println(err)
	}
	return stBytes, sstBytes
}

func (gnb *RanGnbContext) getMccAndMnc() (string, string) {
	return gnb.mcc, gnb.mnc
}

func (gnb *RanGnbContext) getMccAndMncInOctets() []byte {

	// reverse mcc and mnc
	mcc := reverse(gnb.mcc)
	mnc := reverse(gnb.mnc)

	// include mcc and mnc in octets
	oct5 := mcc[1:3]
	var oct6 string
	var oct7 string
	if len(gnb.mnc) == 2 {
		oct6 = "f" + string(mcc[0])
		oct7 = mnc
	} else {
		oct6 = string(mnc[0]) + string(mcc[0])
		oct7 = mnc[1:3]
	}

	// changed for bytes.
	resu, err := hex.DecodeString(oct5 + oct6 + oct7)
	if err != nil {
		fmt.Println(err)
	}

	return resu
}
