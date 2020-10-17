package aper

import (
	"fmt"
	"path"
	"reflect"
	"runtime"
)

type perBitData struct {
	bytes      []byte
	byteOffset uint64
	bitsOffset uint
}

func perTrace(level int, s string) {

	_, file, line, ok := runtime.Caller(1)
	if !ok {
		// logger.AperLog.Debugln(s)
		fmt.Sprintf(s)
	} else {
		// logger.AperLog.Debugf("%s (%s:%d)\n", s, path.Base(file), line)
		fmt.Sprintf(s, path.Base(file), line)
	}
}

func perBitLog(numBits uint64, byteOffset uint64, bitsOffset uint, value interface{}) string {
	if reflect.TypeOf(value).Kind() == reflect.Uint64 {
		return fmt.Sprintf("  [PER got %2d bits, byteOffset(after): %d, bitsOffset(after): %d, value: 0x%0x]",
			numBits, byteOffset, bitsOffset, reflect.ValueOf(value).Uint())
	}
	return fmt.Sprintf("  [PER got %2d bits, byteOffset(after): %d, bitsOffset(after): %d, value: 0x%0x]",
		numBits, byteOffset, bitsOffset, reflect.ValueOf(value).Bytes())

}

// GetBitString is to get BitString with desire size from source byte array with bit offset
func GetBitString(srcBytes []byte, bitsOffset uint, numBits uint) (dstBytes []byte, err error) {
	bitsLeft := uint(len(srcBytes))*8 - bitsOffset
	if numBits > bitsLeft {
		err = fmt.Errorf("Get bits overflow, requireBits: %d, leftBits: %d", numBits, bitsLeft)
		return
	}
	byteLen := (bitsOffset + numBits + 7) >> 3
	numBitsByteLen := (numBits + 7) >> 3
	dstBytes = make([]byte, numBitsByteLen)
	numBitsMask := byte(0xff)
	if modEight := numBits & 0x7; modEight != 0 {
		numBitsMask <<= uint8(8 - (modEight))
	}
	for i := 1; i < int(byteLen); i++ {
		dstBytes[i-1] = srcBytes[i-1]<<bitsOffset | srcBytes[i]>>(8-bitsOffset)
	}
	if byteLen == numBitsByteLen {
		dstBytes[byteLen-1] = srcBytes[byteLen-1] << bitsOffset
	}
	dstBytes[numBitsByteLen-1] &= numBitsMask
	return
}

// GetFewBits is to get Value with desire few bits from source byte with bit offset
// func GetFewBits(srcByte byte, bitsOffset uint, numBits uint) (value uint64, err error) {

// 	if numBits == 0 {
// 		value = 0
// 		return
// 	}
// 	bitsLeft := 8 - bitsOffset
// 	if bitsLeft < numBits {
// 		err = fmt.Errorf("Get bits overflow, requireBits: %d, leftBits: %d", numBits, bitsLeft)
// 		return
// 	}
// 	if bitsOffset == 0 {
// 		value = uint64(srcByte >> (8 - numBits))
// 	} else {
// 		value = uint64((srcByte << bitsOffset) >> (8 - numBits))
// 	}
// 	return
// }

// GetBitsValue is to get Value with desire bits from source byte array with bit offset
func GetBitsValue(srcBytes []byte, bitsOffset uint, numBits uint) (value uint64, err error) {
	var dstBytes []byte
	dstBytes, err = GetBitString(srcBytes, bitsOffset, numBits)
	if err != nil {
		return
	}
	for i, j := 0, numBits; j >= 8; i, j = i+1, j-8 {
		value <<= 8
		value |= uint64(uint(dstBytes[i]))
	}
	if numBitsOff := uint(numBits & 0x7); numBitsOff != 0 {
		var mask uint = (1 << numBitsOff) - 1
		value <<= numBitsOff
		value |= uint64(uint(dstBytes[len(dstBytes)-1]>>(8-numBitsOff)) & mask)
	}
	return
}

func (pd *perBitData) bitCarry() {
	pd.byteOffset += uint64(pd.bitsOffset >> 3)
	pd.bitsOffset = pd.bitsOffset & 0x07
}

func (pd *perBitData) getBitString(numBits uint) (dstBytes []byte, err error) {

	dstBytes, err = GetBitString(pd.bytes[pd.byteOffset:], pd.bitsOffset, numBits)
	if err != nil {
		return
	}
	pd.bitsOffset += uint(numBits)

	pd.bitCarry()
	perTrace(1, perBitLog(uint64(numBits), pd.byteOffset, pd.bitsOffset, dstBytes))
	return
}

func (pd *perBitData) getBitsValue(numBits uint) (value uint64, err error) {
	value, err = GetBitsValue(pd.bytes[pd.byteOffset:], pd.bitsOffset, numBits)
	if err != nil {
		return
	}
	pd.bitsOffset += numBits
	pd.bitCarry()
	perTrace(1, perBitLog(uint64(numBits), pd.byteOffset, pd.bitsOffset, value))
	return
}

func (pd *perBitData) parseAlignBits() error {

	if (pd.bitsOffset & 0x7) > 0 {
		alignBits := 8 - ((pd.bitsOffset) & 0x7)
		perTrace(2, fmt.Sprintf("Aligning %d bits", alignBits))
		if val, err := pd.getBitsValue(alignBits); err != nil {
			return err
		} else if val != 0 {
			return fmt.Errorf("Align Bit is not zero")
		}
	} else if pd.bitsOffset != 0 {
		pd.bitCarry()
	}
	return nil
}

func (pd *perBitData) parseConstraintValue(valueRange int64) (value uint64, err error) {
	perTrace(3, fmt.Sprintf("Getting Constraint Value with range %d", valueRange))

	var bytes uint
	if valueRange <= 255 {
		if valueRange < 0 {
			err = fmt.Errorf("Value range is negative")
			return
		}
		var i uint
		// 1 ~ 8 bits
		for i = 1; i <= 8; i++ {
			upper := 1 << i
			if int64(upper) >= valueRange {
				break
			}
		}
		value, err = pd.getBitsValue(i)
		return
	} else if valueRange == 256 {
		bytes = 1
	} else if valueRange <= 65536 {
		bytes = 2
	} else {
		err = fmt.Errorf("Constraint Value is large than 65536")
		return
	}
	if err = pd.parseAlignBits(); err != nil {
		return
	}
	value, err = pd.getBitsValue(bytes * 8)
	return
}

func (pd *perBitData) parseLength(sizeRange int64, repeat *bool) (value uint64, err error) {
	*repeat = false
	if sizeRange <= 65536 && sizeRange > 0 {
		return pd.parseConstraintValue(sizeRange)
	}

	if err = pd.parseAlignBits(); err != nil {
		return
	}
	firstByte, err := pd.getBitsValue(8)
	if err != nil {
		return
	}
	if (firstByte & 128) == 0 { // #10.9.3.6
		value = firstByte & 0x7F
		return
	} else if (firstByte & 64) == 0 { // #10.9.3.7
		var secondByte uint64
		if secondByte, err = pd.getBitsValue(8); err != nil {
			return
		}
		value = ((firstByte & 63) << 8) | secondByte
		return
	}
	firstByte &= 63
	if firstByte < 1 || firstByte > 4 {
		err = fmt.Errorf("Parse Length Out of Constraint")
		return
	}
	*repeat = true
	value = 16384 * firstByte
	return
}

func (pd *perBitData) parseBitString(extensed bool, lowerBoundPtr *int64, upperBoundPtr *int64) (bitString BitString, err error) {
	var lb, ub, sizeRange int64 = 0, -1, -1
	if !extensed {
		if lowerBoundPtr != nil {
			lb = *lowerBoundPtr
		}
		if upperBoundPtr != nil {
			ub = *upperBoundPtr
			sizeRange = ub - lb + 1
		}
	}
	if ub > 65535 {
		sizeRange = -1
	}
	// initailization
	bitString = BitString{[]byte{}, 0}
	// lowerbound == upperbound
	if sizeRange == 1 {
		sizes := uint64(ub+7) >> 3
		bitString.BitLength = uint64(ub)
		perTrace(2, fmt.Sprintf("Decoding BIT STRING size %d", ub))
		if sizes > 2 {
			if err = pd.parseAlignBits(); err != nil {
				return
			}
			if (pd.byteOffset + sizes) > uint64(len(pd.bytes)) {
				err = fmt.Errorf("PER data out of range")
				return
			}
			bitString.Bytes = pd.bytes[pd.byteOffset : pd.byteOffset+sizes]
			pd.byteOffset += sizes
			pd.bitsOffset = uint(ub & 0x7)
			if pd.bitsOffset > 0 {
				pd.byteOffset--
			}
			perTrace(1, perBitLog(uint64(ub), pd.byteOffset, pd.bitsOffset, bitString.Bytes))
		} else {
			bitString.Bytes, err = pd.getBitString(uint(ub))
		}
		perTrace(2, fmt.Sprintf("Decoded BIT STRING (length = %d): %0.8b", ub, bitString.Bytes))
		return

	}
	repeat := false
	for {
		var rawLength uint64
		if rawLength, err = pd.parseLength(sizeRange, &repeat); err != nil {
			return
		}
		rawLength += uint64(lb)
		perTrace(2, fmt.Sprintf("Decoding BIT STRING size %d", rawLength))
		if rawLength == 0 {
			return
		}
		sizes := (rawLength + 7) >> 3
		if err = pd.parseAlignBits(); err != nil {
			return
		}

		if (pd.byteOffset + sizes) > uint64(len(pd.bytes)) {
			err = fmt.Errorf("PER data out of range")
			return
		}
		bitString.Bytes = append(bitString.Bytes, pd.bytes[pd.byteOffset:pd.byteOffset+sizes]...)
		bitString.BitLength += rawLength
		pd.byteOffset += sizes
		pd.bitsOffset = uint(rawLength & 0x7)
		if pd.bitsOffset != 0 {
			pd.byteOffset--
		}
		perTrace(1, perBitLog(rawLength, pd.byteOffset, pd.bitsOffset, bitString.Bytes))
		perTrace(2, fmt.Sprintf("Decoded BIT STRING (length = %d): %0.8b", rawLength, bitString.Bytes))

		if !repeat {
			// if err = pd.parseAlignBits(); err != nil {
			// 	return
			// }
			break
		}
	}
	return
}
func (pd *perBitData) parseOctetString(extensed bool, lowerBoundPtr *int64, upperBoundPtr *int64) (octetString OctetString, err error) {
	var lb, ub, sizeRange int64 = 0, -1, -1
	if !extensed {
		if lowerBoundPtr != nil {
			lb = *lowerBoundPtr
		}
		if upperBoundPtr != nil {
			ub = *upperBoundPtr
			sizeRange = ub - lb + 1
		}
	}
	if ub > 65535 {
		sizeRange = -1
	}
	// initailization
	octetString = OctetString("")
	// lowerbound == upperbound
	if sizeRange == 1 {
		perTrace(2, fmt.Sprintf("Decoding OCTET STRING size %d", ub))
		if ub > 2 {
			unsignedUB := uint64(ub)
			if err = pd.parseAlignBits(); err != nil {
				return
			}
			if (int64(pd.byteOffset) + ub) > int64(len(pd.bytes)) {
				err = fmt.Errorf("per data out of range")
				return
			}
			octetString = pd.bytes[pd.byteOffset : pd.byteOffset+unsignedUB]
			pd.byteOffset += uint64(ub)
			perTrace(1, perBitLog(8*unsignedUB, pd.byteOffset, pd.bitsOffset, octetString))
		} else {
			octetString, err = pd.getBitString(uint(ub * 8))
		}
		perTrace(2, fmt.Sprintf("Decoded OCTET STRING (length = %d): 0x%0x", ub, octetString))
		return

	}
	repeat := false
	for {
		var rawLength uint64
		if rawLength, err = pd.parseLength(sizeRange, &repeat); err != nil {
			return
		}
		rawLength += uint64(lb)
		perTrace(2, fmt.Sprintf("Decoding OCTET STRING size %d", rawLength))
		if rawLength == 0 {
			return
		} else if err = pd.parseAlignBits(); err != nil {
			return
		}
		if (rawLength + pd.byteOffset) > uint64(len(pd.bytes)) {
			err = fmt.Errorf("per data out of range ")
			return
		}
		octetString = append(octetString, pd.bytes[pd.byteOffset:pd.byteOffset+rawLength]...)
		pd.byteOffset += rawLength
		perTrace(1, perBitLog(8*rawLength, pd.byteOffset, pd.bitsOffset, octetString))
		perTrace(2, fmt.Sprintf("Decoded OCTET STRING (length = %d): 0x%0x", rawLength, octetString))
		if !repeat {
			// if err = pd.parseAlignBits(); err != nil {
			// 	return
			// }
			break
		}
	}
	return
}

func (pd *perBitData) parseBool() (value bool, err error) {
	perTrace(3, fmt.Sprintf("Decoding BOOLEAN Value"))
	bit, err1 := pd.getBitsValue(1)
	if err1 != nil {
		err = err1
		return
	}
	if bit == 1 {
		value = true
		perTrace(2, fmt.Sprintf("Decoded BOOLEAN Value : ture"))
	} else {
		value = false
		perTrace(2, fmt.Sprintf("Decoded BOOLEAN Value : false"))
	}
	return
}

func (pd *perBitData) parseInteger(extensed bool, lowerBoundPtr *int64, upperBoundPtr *int64) (value int64, err error) {
	var lb, ub, valueRange int64 = 0, -1, 0
	if !extensed {
		if lowerBoundPtr == nil {
			perTrace(3, fmt.Sprintf("Decoding INTEGER with Unconstraint Value"))
			valueRange = -1
		} else {
			lb = *lowerBoundPtr
			if upperBoundPtr != nil {
				ub = *upperBoundPtr
				valueRange = ub - lb + 1
				perTrace(3, fmt.Sprintf("Decoding INTEGER with Value Range(%d..%d)", lb, ub))
			} else {
				perTrace(3, fmt.Sprintf("Decoding INTEGER with Semi-Constraint Range(%d..)", lb))
			}
		}
	} else {
		valueRange = -1
		perTrace(3, fmt.Sprintf("Decoding INTEGER with Extensive Value"))
	}
	var rawLength uint
	if valueRange == 1 {
		value = ub
		return
	} else if valueRange <= 0 {
		// semi-constraint or unconstraint
		if err = pd.parseAlignBits(); err != nil {
			return
		}
		if pd.byteOffset >= uint64(len(pd.bytes)) {
			err = fmt.Errorf("per data out of range")
			return
		}
		rawLength = uint(pd.bytes[pd.byteOffset])
		pd.byteOffset++
		perTrace(1, perBitLog(8, pd.byteOffset, pd.bitsOffset, uint64(rawLength)))
	} else if valueRange <= 65536 {
		rawValue, err1 := pd.parseConstraintValue(valueRange)
		if err1 != nil {
			err = err1
		} else {
			value = int64(rawValue) + lb
		}
		return
	} else {
		// valueRange > 65536
		var byteLen uint
		unsignedValueRange := uint64(valueRange - 1)
		for byteLen = 1; byteLen <= 127; byteLen++ {
			unsignedValueRange >>= 8
			if unsignedValueRange == 0 {
				break
			}
		}
		var i, upper uint
		// 1 ~ 8 bits
		for i = 1; i <= 8; i++ {
			upper = 1 << i
			if upper >= byteLen {
				break
			}
		}
		if tempLength, err1 := pd.getBitsValue(i); err1 != nil {
			err = err1
			return
		} else {
			rawLength = uint(tempLength)
		}
		rawLength++
		if err = pd.parseAlignBits(); err != nil {
			return
		}
	}
	perTrace(2, fmt.Sprintf("Decoding INTEGER Length with %d bytes", rawLength))
	var rawValue uint64
	if rawValue, err = pd.getBitsValue(rawLength * 8); err != nil {
		return
	} else if valueRange < 0 {
		signedBitMask := uint64(1 << (rawLength*8 - 1))
		valueMask := signedBitMask - 1
		// negative
		if rawValue&signedBitMask > 0 {
			value = int64((^rawValue)&valueMask+1) * -1
			return
		}
	}
	value = int64(rawValue) + lb
	return

}

// parse ENUMERATED type but do not implement extensive value and different value with index
func (pd *perBitData) parseEnumerated(extensed bool, lowerBoundPtr *int64, upperBoundPtr *int64) (value uint64, err error) {
	if extensed {
		err = fmt.Errorf("Unsupport the extensive value of ENUMERATED ")
		return
	}
	if lowerBoundPtr == nil || upperBoundPtr == nil {
		err = fmt.Errorf("ENUMERATED value constraint is error ")
		return
	}
	lb, ub := *lowerBoundPtr, *upperBoundPtr
	valueRange := ub - lb + 1
	perTrace(2, fmt.Sprintf("Decoding ENUMERATED with Value Range(%d..%d)", lb, ub))
	if valueRange > 1 {
		value, err = pd.parseConstraintValue(valueRange)
	}
	perTrace(2, fmt.Sprintf("Decoded ENUMERATED Value : %d", value))
	return

}
func (pd *perBitData) parseSequenceOf(sizeExtensed bool, params fieldParameters, sliceType reflect.Type) (sliceContent reflect.Value, err error) {
	var lb int64 = 0
	var sizeRange int64
	if params.sizeLowerBound != nil && *params.sizeLowerBound < 65536 {
		lb = *params.sizeLowerBound
	}
	if !sizeExtensed && params.sizeUpperBound != nil && *params.sizeUpperBound < 65536 {
		ub := *params.sizeUpperBound
		sizeRange = ub - lb + 1
		perTrace(3, fmt.Sprintf("Decoding Length of \"SEQUENCE OF\"  with Size Range(%d..%d)", lb, ub))
	} else {
		sizeRange = -1
		perTrace(3, fmt.Sprintf("Decoding Length of \"SEQUENCE OF\" with Semi-Constraint Range(%d..)", lb))
	}

	var numElements uint64
	if sizeRange > 1 {
		numElements, err = pd.parseConstraintValue(sizeRange)
		numElements += uint64(lb)
	} else if sizeRange == 1 {
		numElements += uint64(lb)
	} else {
		if err = pd.parseAlignBits(); err != nil {
			return
		}
		if pd.byteOffset >= uint64(len(pd.bytes)) {
			err = fmt.Errorf("per data out of range")
			return
		}
		numElements = uint64(pd.bytes[pd.byteOffset])
		pd.byteOffset++
		perTrace(1, perBitLog(8, pd.byteOffset, pd.bitsOffset, numElements))
	}
	perTrace(2, fmt.Sprintf("Decoding  \"SEQUENCE OF\" struct %s with len(%d)", sliceType.Elem().Name(), numElements))
	params.sizeExtensible = false
	params.sizeUpperBound = nil
	params.sizeLowerBound = nil
	intNumElements := int(numElements)
	sliceContent = reflect.MakeSlice(sliceType, intNumElements, intNumElements)
	for i := 0; i < intNumElements; i++ {
		err = parseField(sliceContent.Index(i), pd, params)
		if err != nil {
			return
		}
	}
	return
}

func (pd *perBitData) getChoiceIndex(extensed bool, upperBoundPtr *int64) (present int, err error) {
	if extensed {
		err = fmt.Errorf("Unsupport value of CHOICE type is in Extensed")
	} else if upperBoundPtr == nil {
		err = fmt.Errorf("The upper bound of CHIOCE is missing")
	} else if ub := *upperBoundPtr; ub < 0 {
		err = fmt.Errorf("The upper bound of CHIOCE is negative")
	} else if rawChoice, err1 := pd.parseConstraintValue(ub + 1); err1 != nil {
		err = err1
	} else {
		perTrace(2, fmt.Sprintf("Decoded Present index of CHOICE is %d + 1", rawChoice))
		present = int(rawChoice) + 1
	}
	return
}
func getReferenceFieldValue(v reflect.Value) (value int64, err error) {
	fieldType := v.Type()
	switch v.Kind() {
	case reflect.Int, reflect.Int32, reflect.Int64:
		value = v.Int()
	case reflect.Struct:
		if fieldType.Field(0).Name == "Present" {
			present := int(v.Field(0).Int())
			if present == 0 {
				err = fmt.Errorf("ReferenceField Value present is 0(present's field number)")
			} else if present >= fieldType.NumField() {
				err = fmt.Errorf("Present is bigger than number of struct field")
			} else {
				value, err = getReferenceFieldValue(v.Field(present))
			}
		} else {
			value, err = getReferenceFieldValue(v.Field(0))
		}
	default:
		err = fmt.Errorf("OpenType reference only support INTEGER")
	}
	return
}

func (pd *perBitData) parseOpenType(v reflect.Value, params fieldParameters) (err error) {

	pdOpenType := &perBitData{[]byte(""), 0, 0}
	repeat := false
	for {
		var rawLength uint64
		if rawLength, err = pd.parseLength(-1, &repeat); err != nil {
			return
		}
		if rawLength == 0 {
			break
		} else if err = pd.parseAlignBits(); err != nil {
			return
		}
		if (rawLength + pd.byteOffset) > uint64(len(pd.bytes)) {
			err = fmt.Errorf("per data out of range ")
			return
		}
		pdOpenType.bytes = append(pdOpenType.bytes, pd.bytes[pd.byteOffset:pd.byteOffset+rawLength]...)
		pd.byteOffset += rawLength

		if !repeat {
			if err = pd.parseAlignBits(); err != nil {
				return
			}
			break
		}
	}
	perTrace(2, fmt.Sprintf("Decoding OpenType %s with (len = %d byte)", v.Type().String(), len(pdOpenType.bytes)))
	err = parseField(v, pdOpenType, params)
	perTrace(2, fmt.Sprintf("Decoded OpenType %s", v.Type().String()))
	return
}

// parseField is the main parsing function. Given a byte slice and an offset
// into the array, it will try to parse a suitable ASN.1 value out and store it
// in the given Value. TODO : ObjectIdenfier, handle extension Field
func parseField(v reflect.Value, pd *perBitData, params fieldParameters) (err error) {
	fieldType := v.Type()

	// If we have run out of data return error.
	if pd.byteOffset == uint64(len(pd.bytes)) {
		err = fmt.Errorf("sequence truncated")
		return
	}
	if v.Kind() == reflect.Ptr {
		ptr := reflect.New(fieldType.Elem())
		v.Set(ptr)
		err = parseField(v.Elem(), pd, params)
		return
	}
	sizeExtensible := false
	valueExtensible := false
	if params.sizeExtensible {
		if bitsValue, err1 := pd.getBitsValue(1); err1 != nil {
			return err1
		} else if bitsValue != 0 {
			sizeExtensible = true
		}
		perTrace(2, fmt.Sprintf("Decoded Size Extensive Bit : %t", sizeExtensible))
	}
	if params.valueExtensible && v.Kind() != reflect.Slice {
		if bitsValue, err1 := pd.getBitsValue(1); err1 != nil {
			return err1
		} else if bitsValue != 0 {
			valueExtensible = true
		}
		perTrace(2, fmt.Sprintf("Decoded Value Extensive Bit : %t", valueExtensible))
	}

	// We deal with the structures defined in this package first.
	switch fieldType {
	case BitStringType:
		bitString, err1 := pd.parseBitString(sizeExtensible, params.sizeLowerBound, params.sizeUpperBound)

		if err1 != nil {
			return err1
		}
		v.Set(reflect.ValueOf(bitString))
		return
	case ObjectIdentifierType:
		err = fmt.Errorf("Unsupport ObjectIdenfier type")
		return
	case OctetStringType:
		octetString, err1 := pd.parseOctetString(sizeExtensible, params.sizeLowerBound, params.sizeUpperBound)
		if err1 == nil {
			v.Set(reflect.ValueOf(octetString))
		}
		err = err1
		return
	case EnumeratedType:
		parsedEnum, err1 := pd.parseEnumerated(valueExtensible, params.valueLowerBound, params.valueUpperBound)
		if err1 == nil {
			v.SetUint(parsedEnum)
		}
		err = err1
		return
	}
	switch val := v; val.Kind() {
	case reflect.Bool:
		parsedBool, err1 := pd.parseBool()
		if err1 == nil {
			val.SetBool(parsedBool)
		}
		err = err1
		return
	case reflect.Int, reflect.Int32, reflect.Int64:
		parsedInt, err1 := pd.parseInteger(valueExtensible, params.valueLowerBound, params.valueUpperBound)
		if err1 == nil {
			val.SetInt(parsedInt)
			perTrace(2, fmt.Sprintf("Decoded INTEGER Value : %d", parsedInt))
		}
		err = err1
		return

	case reflect.Struct:

		structType := fieldType
		var structParams []fieldParameters
		var optionalCount uint
		var optionalPresents uint64

		// pass tag for optional
		for i := 0; i < structType.NumField(); i++ {
			if structType.Field(i).PkgPath != "" {
				err = fmt.Errorf("struct contains unexported fields : " + structType.Field(i).PkgPath)
				return
			}
			tempParams := parseFieldParameters(structType.Field(i).Tag.Get("aper"))
			// for optional flag
			if tempParams.optional {
				optionalCount++
			}
			structParams = append(structParams, tempParams)
		}

		if optionalCount > 0 {
			if optionalPresents, err = pd.getBitsValue(optionalCount); err != nil {
				return
			}
			perTrace(2, fmt.Sprintf("optionalPresents is %0b", optionalPresents))
		}

		// CHOICE or OpenType
		if structType.NumField() > 0 && structType.Field(0).Name == "Present" {
			var present int = 0
			if params.openType {
				if params.referenceFieldValue == nil {
					err = fmt.Errorf("OpenType reference value is empty")
					return
				}
				refValue := *params.referenceFieldValue

				for j, param := range structParams {
					if j == 0 {
						continue
					}
					if param.referenceFieldValue != nil && *param.referenceFieldValue == refValue {
						present = j
						break
					}
				}
				if present == 0 {
					err = fmt.Errorf("OpenType reference value does not match any field")
				} else if present >= structType.NumField() {
					err = fmt.Errorf("OpenType Present is bigger than number of struct field")
				} else {
					val.Field(0).SetInt(int64(present))
					perTrace(2, fmt.Sprintf("Decoded Present index of OpenType is %d ", present))
					err = pd.parseOpenType(val.Field(present), structParams[present])
				}
			} else {
				present, err = pd.getChoiceIndex(valueExtensible, params.valueUpperBound)
				if err != nil {
					// logger.AperLog.Errorf("pd.getChoiceIndex Error")
				}
				val.Field(0).SetInt(int64(present))
				if present == 0 {
					err = fmt.Errorf("CHOICE present is 0(present's field number)")
				} else if present >= structType.NumField() {
					err = fmt.Errorf("CHOICE Present is bigger than number of struct field")
				} else {
					err = parseField(val.Field(present), pd, structParams[present])
				}
			}
			return

		}

		for i := 0; i < structType.NumField(); i++ {
			if structParams[i].optional && optionalCount > 0 {
				optionalCount--
				if optionalPresents&(1<<optionalCount) == 0 {
					perTrace(3, fmt.Sprintf("Field \"%s\" in %s is OPTIONAL and not present", structType.Field(i).Name, structType))
					continue
				} else {
					perTrace(3, fmt.Sprintf("Field \"%s\" in %s is OPTIONAL and present", structType.Field(i).Name, structType))
				}
			}
			// for open type reference
			if structParams[i].openType {
				fieldName := structParams[i].referenceFieldName
				var index int
				for index = 0; index < i; index++ {
					if structType.Field(index).Name == fieldName {
						break
					}
				}
				if index == i {
					err = fmt.Errorf("Open type is not reference to the other field in the struct")
					return
				}
				structParams[i].referenceFieldValue = new(int64)
				*structParams[i].referenceFieldValue, err = getReferenceFieldValue(val.Field(index))
				if err != nil {
					return
				}
			}
			err = parseField(val.Field(i), pd, structParams[i])
			if err != nil {
				return
			}
		}
		return
	case reflect.Slice:
		sliceType := fieldType
		newSlice, err1 := pd.parseSequenceOf(sizeExtensible, params, sliceType)
		if err1 == nil {
			val.Set(newSlice)
		}
		err = err1
		return
	case reflect.String:
		perTrace(2, fmt.Sprintf("Decoding PrintableString using Octet String decoding method"))

		octetString, err1 := pd.parseOctetString(sizeExtensible, params.sizeLowerBound, params.sizeUpperBound)
		err = err1
		if err1 == nil {
			printableString := string(octetString)
			val.SetString(printableString)
			perTrace(2, fmt.Sprintf("Decoded PrintableString : \"%s\"", printableString))
		}
		return

	}
	err = fmt.Errorf("unsupported: " + v.Type().String())
	return
}

// Unmarshal parses the APER-encoded ASN.1 data structure b
// and uses the reflect package to fill in an arbitrary value pointed at by value.
// Because Unmarshal uses the reflect package, the structs
// being written to must use upper case field names.
//
// An ASN.1 INTEGER can be written to an int, int32, int64,
// If the encoded value does not fit in the Go type,
// Unmarshal returns a parse error.
//
// An ASN.1 BIT STRING can be written to a BitString.
//
// An ASN.1 OCTET STRING can be written to a []byte.
//
// An ASN.1 OBJECT IDENTIFIER can be written to an
// ObjectIdentifier.
//
// An ASN.1 ENUMERATED can be written to an Enumerated.
//
// Any of the above ASN.1 values can be written to an interface{}.
// The value stored in the interface has the corresponding Go type.
// For integers, that type is int64.
//
// An ASN.1 SEQUENCE OF x can be written
// to a slice if an x can be written to the slice's element type.
//
// An ASN.1 SEQUENCE can be written to a struct
// if each of the elements in the sequence can be
// written to the corresponding element in the struct.
//
// The following tags on struct fields have special meaning to Unmarshal:
//
//	optional        	OPTIONAL tag in SEQUENCE
//	sizeExt             specifies that size  is extensible
//	valueExt            specifies that value is extensible
//	sizeLB		        set the minimum value of size constraint
//	sizeUB              set the maximum value of value constraint
//	valueLB		        set the minimum value of size constraint
//	valueUB             set the maximum value of value constraint
//	default             sets the default value
//	openType            specifies the open Type
//  referenceFieldName	the string of the reference field for this type (only if openType used)
//  referenceFieldValue	the corresponding value of the reference field for this type (only if openType used)
//
// Other ASN.1 types are not supported; if it encounters them,
// Unmarshal returns a parse error.
func Unmarshal(b []byte, value interface{}) error {
	return UnmarshalWithParams(b, value, "")
}

// UnmarshalWithParams allows field parameters to be specified for the
// top-level element. The form of the params is the same as the field tags.
func UnmarshalWithParams(b []byte, value interface{}, params string) error {
	v := reflect.ValueOf(value).Elem()
	pd := &perBitData{b, 0, 0}
	return parseField(v, pd, parseFieldParameters(params))

}
