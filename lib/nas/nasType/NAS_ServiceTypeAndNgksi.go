package nasType

// ServiceTypeAndNgksi 9.11.3.32 9.11.3.50
// ServiceTypeValue Row, sBit, len = [0, 0], 8 , 4
// TSC Row, sBit, len = [0, 0], 4 , 1
// NasKeySetIdentifiler Row, sBit, len = [0, 0], 3 , 3
type ServiceTypeAndNgksi struct {
	Octet uint8
}

func NewServiceTypeAndNgksi() (serviceTypeAndNgksi *ServiceTypeAndNgksi) {
	serviceTypeAndNgksi = &ServiceTypeAndNgksi{}
	return serviceTypeAndNgksi
}

// ServiceTypeAndNgksi 9.11.3.32 9.11.3.50
// ServiceTypeValue Row, sBit, len = [0, 0], 8 , 4
func (a *ServiceTypeAndNgksi) GetServiceTypeValue() (serviceTypeValue uint8) {
	return a.Octet & GetBitMask(8, 4) >> (4)
}

// ServiceTypeAndNgksi 9.11.3.32 9.11.3.50
// ServiceTypeValue Row, sBit, len = [0, 0], 8 , 4
func (a *ServiceTypeAndNgksi) SetServiceTypeValue(serviceTypeValue uint8) {
	a.Octet = (a.Octet & 15) + ((serviceTypeValue & 15) << 4)
}

// ServiceTypeAndNgksi 9.11.3.32 9.11.3.50
// TSC Row, sBit, len = [0, 0], 4 , 1
func (a *ServiceTypeAndNgksi) GetTSC() (tSC uint8) {
	return a.Octet & GetBitMask(4, 3) >> (3)
}

// ServiceTypeAndNgksi 9.11.3.32 9.11.3.50
// TSC Row, sBit, len = [0, 0], 4 , 1
func (a *ServiceTypeAndNgksi) SetTSC(tSC uint8) {
	a.Octet = (a.Octet & 247) + ((tSC & 1) << 3)
}

// ServiceTypeAndNgksi 9.11.3.32 9.11.3.50
// NasKeySetIdentifiler Row, sBit, len = [0, 0], 3 , 3
func (a *ServiceTypeAndNgksi) GetNasKeySetIdentifiler() (nasKeySetIdentifiler uint8) {
	return a.Octet & GetBitMask(3, 0)
}

// ServiceTypeAndNgksi 9.11.3.32 9.11.3.50
// NasKeySetIdentifiler Row, sBit, len = [0, 0], 3 , 3
func (a *ServiceTypeAndNgksi) SetNasKeySetIdentifiler(nasKeySetIdentifiler uint8) {
	a.Octet = (a.Octet & 248) + (nasKeySetIdentifiler & 7)
}
