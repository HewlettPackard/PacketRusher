package procedures

type UeTesterMessageType int32

const (
	Registration      UeTesterMessageType = 0
	Deregistration    UeTesterMessageType = 1
	NewPDUSession     UeTesterMessageType = 2
	DestroyPDUSession UeTesterMessageType = 3
	Terminate         UeTesterMessageType = 4
	Kill              UeTesterMessageType = 5
)

type UeTesterMessage struct {
	Type UeTesterMessageType
	Param uint8
}