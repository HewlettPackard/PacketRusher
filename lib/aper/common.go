package aper

import (
	"strconv"
	"strings"
)

// fieldParameters is the parsed representation of tag string from a structure field.
type fieldParameters struct {
	optional            bool   // true iff the type has OPTIONAL tag.
	sizeExtensible      bool   // true iff the size can be extensed.
	valueExtensible     bool   // true iff the value can be extensed.
	sizeLowerBound      *int64 // a sizeLowerBound is the minimum size of type constraint(maybe nil).
	sizeUpperBound      *int64 // a sizeUpperBound is the maximum size of type constraint(maybe nil).
	valueLowerBound     *int64 // a valueLowerBound is the minimum value of type constraint(maybe nil).
	valueUpperBound     *int64 // a valueUpperBound is the maximum value of type constraint(maybe nil).
	defaultValue        *int64 // a default value for INTEGER and ENUMERATED typed fields (maybe nil).
	openType            bool   // true iff this type is opentype.
	referenceFieldName  string // the field to get to get the corresrponding value of this type(maybe nil).
	referenceFieldValue *int64 // the field value which map to this type(maybe nil).
}

// Given a tag string with the format specified in the package comment,
// parseFieldParameters will parse it into a fieldParameters structure,
// ignoring unknown parts of the string. TODO:PrintableString
func parseFieldParameters(str string) (params fieldParameters) {
	for _, part := range strings.Split(str, ",") {
		switch {
		case part == "optional":
			params.optional = true
		case part == "sizeExt":
			params.sizeExtensible = true
		case part == "valueExt":
			params.valueExtensible = true
		case strings.HasPrefix(part, "sizeLB:"):
			i, err := strconv.ParseInt(part[7:], 10, 64)
			if err == nil {
				params.sizeLowerBound = new(int64)
				*params.sizeLowerBound = i
			}
		case strings.HasPrefix(part, "sizeUB:"):
			i, err := strconv.ParseInt(part[7:], 10, 64)
			if err == nil {
				params.sizeUpperBound = new(int64)
				*params.sizeUpperBound = i
			}
		case strings.HasPrefix(part, "valueLB:"):
			i, err := strconv.ParseInt(part[8:], 10, 64)
			if err == nil {
				params.valueLowerBound = new(int64)
				*params.valueLowerBound = i
			}
		case strings.HasPrefix(part, "valueUB:"):
			i, err := strconv.ParseInt(part[8:], 10, 64)
			if err == nil {
				params.valueUpperBound = new(int64)
				*params.valueUpperBound = i
			}
		case strings.HasPrefix(part, "default:"):
			i, err := strconv.ParseInt(part[8:], 10, 64)
			if err == nil {
				params.defaultValue = new(int64)
				*params.defaultValue = i
			}
		case part == "openType":
			params.openType = true
		case strings.HasPrefix(part, "referenceFieldName:"):
			params.referenceFieldName = part[19:]
		case strings.HasPrefix(part, "referenceFieldValue:"):
			i, err := strconv.ParseInt(part[20:], 10, 64)
			if err == nil {
				params.referenceFieldValue = new(int64)
				*params.referenceFieldValue = i
			}
		}
	}
	return
}
