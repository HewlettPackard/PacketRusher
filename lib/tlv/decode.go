package tlv

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

func Unmarshal(b []byte, v interface{}) error {
	return decodeValue(b, v)
}

func decodeValue(b []byte, v interface{}) (err error) {
	value := reflect.ValueOf(v)

	if unmarshaler, ok := value.Interface().(encoding.BinaryUnmarshaler); ok {
		err := unmarshaler.UnmarshalBinary(b)
		return err
	}

	value = reflect.Indirect(value)
	valueType := reflect.TypeOf(value.Interface())
	switch value.Kind() {
	case reflect.Int8:
		var tmp = int64(int8(b[0]))
		value.SetInt(tmp)
	case reflect.Int16:
		var tmp = int64(int16(binary.BigEndian.Uint16(b)))
		value.SetInt(tmp)
	case reflect.Int32:
		var tmp = int64(int32(binary.BigEndian.Uint32(b)))
		value.SetInt(tmp)
	case reflect.Int64:
		var tmp = int64(binary.BigEndian.Uint64(b))
		value.SetInt(tmp)
	case reflect.Int:
		var tmp = int64(binary.BigEndian.Uint64(b))
		value.SetInt(tmp)
	case reflect.Uint8:
		var tmp = uint64(b[0])
		value.SetUint(tmp)
	case reflect.Uint16:
		var tmp = uint64(binary.BigEndian.Uint16(b))
		value.SetUint(tmp)
	case reflect.Uint32:
		var tmp = uint64(binary.BigEndian.Uint32(b))
		value.SetUint(tmp)
	case reflect.Uint64:
		var tmp = binary.BigEndian.Uint64(b)
		value.SetUint(tmp)
	case reflect.Uint:
		var tmp = binary.BigEndian.Uint64(b)
		value.SetUint(tmp)
	case reflect.String:
		value.SetString(string(b))
	case reflect.Struct:
		tlvFragment, _ := parseTLV(b)
		for i := 0; i < value.NumField(); i++ {
			fieldValue := value.Field(i)
			fieldType := valueType.Field(i)

			tag, hasTLV := fieldType.Tag.Lookup("tlv")
			if !hasTLV {
				return errors.New("field " + fieldType.Name + " need tag `tlv`")
			}

			tagVal, err := strconv.Atoi(tag)
			if err != nil {
				return fmt.Errorf("invalid tlv tag \"%s\", need to be decimal number", tag)
			}

			if len(tlvFragment[tagVal]) == 0 {
				continue
			}

			if fieldValue.Kind() == reflect.Ptr && fieldValue.IsNil() {
				fieldValue.Set(reflect.New(fieldValue.Type().Elem()))
			} else if fieldValue.Kind() == reflect.Slice && fieldValue.IsNil() {
				fieldValue.Set(reflect.MakeSlice(fieldValue.Type(), 0, 1))
			}
			for _, buf := range tlvFragment[tagVal] {
				if fieldValue.Kind() != reflect.Ptr {
					fieldValue = fieldValue.Addr()
				}
				err = decodeValue(buf, fieldValue.Interface())
				if err != nil {
					return err
				}
			}
		}
	case reflect.Slice:
		if valueType.Elem().Kind() == reflect.Uint8 {
			value.SetBytes(b)
		} else if valueType.Elem().Kind() == reflect.Ptr || valueType.Elem().Kind() == reflect.Struct || isNumber(valueType.Elem()) {
			elemValue := reflect.New(valueType.Elem())
			_ = decodeValue(b, elemValue.Interface())
			value.Set(reflect.Append(value, elemValue.Elem()))
		} else {
			return errors.New("value type `Slice of " + valueType.String() + "` is not support decode")
		}
	}
	return nil
}

func parseTLV(b []byte) (fragments, error) {
	tlvFragment := make(fragments)
	buffer := bytes.NewBuffer(b)

	var tag uint16
	var length uint16
	for {
		if err := binary.Read(buffer, binary.BigEndian, &tag); err != nil {
			fmt.Printf("Binary Read error: %v", err)
		}
		if err := binary.Read(buffer, binary.BigEndian, &length); err != nil {
			fmt.Printf("Binary Read error: %v", err)
		}
		value := make([]byte, length)
		if _, err := buffer.Read(value); err != nil {
			return nil, err
		}
		tlvFragment.Add(int(tag), value)
		if buffer.Len() == 0 {
			break
		}
	}
	return tlvFragment, nil
}
