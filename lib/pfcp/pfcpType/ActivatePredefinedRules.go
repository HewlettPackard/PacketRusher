package pfcpType

type ActivatePredefinedRules struct {
	PredefinedRulesName []byte
}

func (a *ActivatePredefinedRules) MarshalBinary() (data []byte, err error) {
	// Octet 5 to (n+4)
	data = a.PredefinedRulesName

	return data, nil
}

func (a *ActivatePredefinedRules) UnmarshalBinary(data []byte) error {
	// Octet 5 to (n+4)
	a.PredefinedRulesName = data

	return nil
}
