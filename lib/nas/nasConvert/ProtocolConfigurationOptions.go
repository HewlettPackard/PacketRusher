package nasConvert

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"my5G-RANTester/lib/nas/nasMessage"
	"net"
)

type ProtocolOrContainerUnit struct {
	ProtocolOrContainerID uint16
	LengthofContents      uint8
	Contents              []byte
}

type ProtocolConfigurationOptions struct {
	ProtocolOrContainerList []*ProtocolOrContainerUnit
}

type PCOReadingState int

const (
	ReadingID PCOReadingState = iota
	ReadingLength
	ReadingContent
)

func NewProtocolOrContainerUnit() (pcu *ProtocolOrContainerUnit) {
	pcu = &ProtocolOrContainerUnit{
		ProtocolOrContainerID: 0,
		LengthofContents:      0,
		Contents:              []byte{},
	}
	return
}

func NewProtocolConfigurationOptions() (pco *ProtocolConfigurationOptions) {

	pco = &ProtocolConfigurationOptions{
		ProtocolOrContainerList: make([]*ProtocolOrContainerUnit, 0),
	}

	return
}

func (protocolConfigurationOptions *ProtocolConfigurationOptions) Marshal() (nas []byte) {

	var metaInfo uint8
	var extension uint8 = 1
	var spare uint8 = 0
	var configurationProtocol uint8 = 0
	buffer := new(bytes.Buffer)

	metaInfo = (extension << 7) | (spare << 6) | (configurationProtocol)
	binary.Write(buffer, binary.BigEndian, &metaInfo)

	for _, containerUnit := range protocolConfigurationOptions.ProtocolOrContainerList {

		binary.Write(buffer, binary.BigEndian, &containerUnit.ProtocolOrContainerID)
		binary.Write(buffer, binary.BigEndian, &containerUnit.LengthofContents)
		binary.Write(buffer, binary.BigEndian, &containerUnit.Contents)
	}

	nas = buffer.Bytes()
	return
}

func (protocolConfigurationOptions *ProtocolConfigurationOptions) UnMarshal(data []byte) (err error) {
	// logger.ConvertLog.Traceln("In ProtocolConfigurationOptions UnMarshal")

	var Buf uint8
	numOfBytes := len(data)
	byteReader := bytes.NewReader(data)
	err = binary.Read(byteReader, binary.BigEndian, &Buf)
	if err != nil {
		return err
	}

	numOfBytes = numOfBytes - 1
	readingState := ReadingID
	var curContainer *ProtocolOrContainerUnit

	for numOfBytes > 0 {

		switch readingState {
		case ReadingID:
			curContainer = NewProtocolOrContainerUnit()
			err = binary.Read(byteReader, binary.BigEndian, &curContainer.ProtocolOrContainerID)
			if err != nil {
				return err
			}
			// logger.ConvertLog.Traceln("Reading ID: ", strconv.Itoa(int(curContainer.ProtocolOrContainerID)))
			readingState = ReadingLength
			numOfBytes = numOfBytes - 2
		case ReadingLength:
			err = binary.Read(byteReader, binary.BigEndian, &curContainer.LengthofContents)
			if err != nil {
				return err
			}
			// logger.ConvertLog.Traceln("Reading Length: ", strconv.Itoa(int(curContainer.LengthofContents)))
			readingState = ReadingContent
			numOfBytes = numOfBytes - 1
			if curContainer.LengthofContents == 0 {
				protocolConfigurationOptions.ProtocolOrContainerList = append(protocolConfigurationOptions.ProtocolOrContainerList, curContainer)
				// logger.ConvertLog.Traceln("For loop ProtocolOrContainerList: ", protocolConfigurationOptions.ProtocolOrContainerList)
			}
		case ReadingContent:
			if curContainer.LengthofContents > 0 {
				curContainer.Contents = make([]uint8, curContainer.LengthofContents)
				err = binary.Read(byteReader, binary.BigEndian, curContainer.Contents)
				if err != nil {
					return err
				}
				protocolConfigurationOptions.ProtocolOrContainerList = append(protocolConfigurationOptions.ProtocolOrContainerList, curContainer)
				// logger.ConvertLog.Traceln("For loop ProtocolOrContainerList: ", protocolConfigurationOptions.ProtocolOrContainerList)
			}
			numOfBytes = numOfBytes - int(curContainer.LengthofContents)
			readingState = ReadingID
		}
	}

	// logger.ConvertLog.Infoln("ProtocolOrContainerList: ", protocolConfigurationOptions.ProtocolOrContainerList)
	return
}

func (protocolConfigurationOptions *ProtocolConfigurationOptions) AddDNSServerIPv4AddressRequest() {
	protocolOrContainerUnit := NewProtocolOrContainerUnit()

	protocolOrContainerUnit.ProtocolOrContainerID = nasMessage.DNSServerIPv4AddressRequestUL
	protocolOrContainerUnit.LengthofContents = 0

	protocolConfigurationOptions.ProtocolOrContainerList = append(protocolConfigurationOptions.ProtocolOrContainerList, protocolOrContainerUnit)
}

func (protocolConfigurationOptions *ProtocolConfigurationOptions) AddDNSServerIPv6AddressRequest() {
	protocolOrContainerUnit := NewProtocolOrContainerUnit()

	protocolOrContainerUnit.ProtocolOrContainerID = nasMessage.DNSServerIPv6AddressRequestUL
	protocolOrContainerUnit.LengthofContents = 0

	protocolConfigurationOptions.ProtocolOrContainerList = append(protocolConfigurationOptions.ProtocolOrContainerList, protocolOrContainerUnit)
}

func (protocolConfigurationOptions *ProtocolConfigurationOptions) AddIPAddressAllocationViaNASSignallingUL() {
	protocolOrContainerUnit := NewProtocolOrContainerUnit()

	protocolOrContainerUnit.ProtocolOrContainerID = nasMessage.IPAddressAllocationViaNASSignallingUL
	protocolOrContainerUnit.LengthofContents = 0

	protocolConfigurationOptions.ProtocolOrContainerList = append(protocolConfigurationOptions.ProtocolOrContainerList, protocolOrContainerUnit)
}

func (protocolConfigurationOptions *ProtocolConfigurationOptions) AddDNSServerIPv4Address(dnsIP net.IP) (err error) {

	if dnsIP.To4() == nil {
		err = fmt.Errorf("The DNS IP should be IPv4 in AddDNSServerIPv4Address!")
		return
	}
	dnsIP = dnsIP.To4()

	if len(dnsIP) != net.IPv4len {
		err = fmt.Errorf("The length of DNS IPv4 is wrong!")
		return
	}

	// logger.ConvertLog.Traceln("In AddDNSServerIPv4Address")
	protocolOrContainerUnit := NewProtocolOrContainerUnit()

	protocolOrContainerUnit.ProtocolOrContainerID = nasMessage.DNSServerIPv4AddressDL
	protocolOrContainerUnit.LengthofContents = uint8(net.IPv4len)
	// logger.ConvertLog.Traceln("LengthofContents: ", protocolOrContainerUnit.LengthofContents)
	protocolOrContainerUnit.Contents = append(protocolOrContainerUnit.Contents, dnsIP.To4()...)
	// logger.ConvertLog.Traceln("Contents: ", protocolOrContainerUnit.Contents)

	protocolConfigurationOptions.ProtocolOrContainerList = append(protocolConfigurationOptions.ProtocolOrContainerList, protocolOrContainerUnit)
	return
}

func (protocolConfigurationOptions *ProtocolConfigurationOptions) AddDNSServerIPv6Address(dnsIP net.IP) (err error) {

	if dnsIP.To16() == nil {
		err = fmt.Errorf("The DNS IP should be IPv6 in AddDNSServerIPv6Address!")
		return
	}

	if len(dnsIP) != net.IPv6len {
		err = fmt.Errorf("The length of DNS IPv6 is wrong!")
		return
	}

	protocolOrContainerUnit := NewProtocolOrContainerUnit()

	protocolOrContainerUnit.ProtocolOrContainerID = nasMessage.DNSServerIPv6AddressDL
	protocolOrContainerUnit.LengthofContents = uint8(net.IPv6len)
	protocolOrContainerUnit.Contents = append(protocolOrContainerUnit.Contents, dnsIP.To16()...)

	protocolConfigurationOptions.ProtocolOrContainerList = append(protocolConfigurationOptions.ProtocolOrContainerList, protocolOrContainerUnit)
	return
}
