package openapi

import (
	"fmt"
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"
)

func openAPIDecodeHook(from reflect.Type, to reflect.Type, v interface{}) (interface{}, error) {
	// convert OpenAPI DateTime to time.Time based on RFC3339
	if to == reflect.TypeOf(time.Time{}) && from == reflect.TypeOf("") {
		return time.Parse(time.RFC3339, v.(string))
	}
	return v, nil
}

// Convert - convert map[string]interface{} to openapi models
func Convert(from interface{}, to interface{}) error {
	config := mapstructure.DecoderConfig{
		DecodeHook: openAPIDecodeHook,
		Result:     to,
	}

	decoder, err := mapstructure.NewDecoder(&config)
	if err != nil {
		return fmt.Errorf("openapi: converter setup failed: %v", err)
	}

	err = decoder.Decode(from)
	if err != nil {
		return fmt.Errorf("openapi: convert to %v failed: %v", reflect.TypeOf(to), err)
	}

	return nil
}
