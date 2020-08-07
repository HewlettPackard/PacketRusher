package TimeDecode

import (
	"reflect"
	"time"

	"free5gc/lib/openapi/models"

	"github.com/mitchellh/mapstructure"
)

// Decode - Only support []map[string]interface to []models.NfProfile
func Decode(source interface{}, format string) ([]models.NfProfile, error) {
	var target []models.NfProfile

	// config mapstruct
	stringToDateTimeHook := func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if t == reflect.TypeOf(time.Time{}) && f == reflect.TypeOf("") {
			return time.Parse(format, data.(string))
		}
		return data, nil
	}

	config := mapstructure.DecoderConfig{
		DecodeHook: stringToDateTimeHook,
		Result:     &target,
	}

	decoder, err := mapstructure.NewDecoder(&config)
	if err != nil {
		return nil, err
	}

	// Decode result to NfProfile structure
	err = decoder.Decode(source)
	if err != nil {
		return nil, err
	}
	return target, nil
}
