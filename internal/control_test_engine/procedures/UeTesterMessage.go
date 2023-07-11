package procedures

type UeTesterMessage int32

const (
	Registration   UeTesterMessage = 0
	Deregistration UeTesterMessage = 1
	NewPDUSession  UeTesterMessage = 2
	Kill           UeTesterMessage = 3
)