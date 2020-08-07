package openapi

import (
	"encoding/hex"
)

// SupportedFeature - bytes used to indicate the features supported by a API
// that is used as defined in subclause 6.6 in 3GPP TS 29.500
type SupportedFeature []byte

// NewSupportedFeature - new NewSupportedFeature from string
func NewSupportedFeature(suppFeat string) (SupportedFeature, error) {
	// padding for hex decode
	if len(suppFeat)%2 != 0 {
		suppFeat = "0" + suppFeat
	}

	supportedFeature, err := hex.DecodeString(suppFeat)
	return supportedFeature, err
}

// String - convert SupportedFeature to hex format
func (suppoertedFeature SupportedFeature) String() string {
	return hex.EncodeToString(suppoertedFeature)
}

// GetFeature - get nth feature is supported
func (suppoertedFeature SupportedFeature) GetFeature(n int) bool {
	byteIndex := len(suppoertedFeature) - ((n - 1) / 8) - 1
	bitShift := uint8((n - 1) % 8)

	if byteIndex < 0 {
		return false
	}

	if suppoertedFeature[byteIndex]&(0x01<<bitShift) > 0 {
		return true
	}

	return false
}

// NegotiateWith - Negotiate with other supported feature
func (suppoertedFeature SupportedFeature) NegotiateWith(incomingSuppFeat SupportedFeature) SupportedFeature {
	var suppFeatA, suppFeatB, negotiateFeature SupportedFeature
	var negotiatedFeatureLength, lengthDiff int
	// padding short one
	if len(suppoertedFeature) < len(incomingSuppFeat) {
		suppFeatA = incomingSuppFeat
		suppFeatB = make(SupportedFeature, len(incomingSuppFeat))
		lengthDiff = len(incomingSuppFeat) - len(suppoertedFeature)
		copy(suppFeatB[lengthDiff:], suppoertedFeature)
		negotiatedFeatureLength = len(incomingSuppFeat)
	} else {
		suppFeatA = suppoertedFeature
		suppFeatB = make(SupportedFeature, len(suppoertedFeature))
		lengthDiff = len(suppoertedFeature) - len(incomingSuppFeat)
		copy(suppFeatB[lengthDiff:], incomingSuppFeat)
		negotiatedFeatureLength = len(suppoertedFeature)
	}

	negotiateFeature = make(SupportedFeature, negotiatedFeatureLength)

	for i := 0; i < negotiatedFeatureLength; i++ {
		negotiateFeature[i] = suppFeatA[i] & suppFeatB[i]
	}

	return negotiateFeature
}
