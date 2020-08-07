package pfcp

import (
	"free5gc/lib/pfcp/logger"
	"net"
	"time"
)

type TransactionType uint8

type TxTable map[uint32]*Transaction

const (
	SendingRequest TransactionType = iota
	SendingResponse
)

const (
	NumOfResend                 = 3
	ResendRequestTimeOutPeriod  = 3
	ResendResponseTimeOutPeriod = 15
)

// Transaction - represent the transaction state of pfcp message
type Transaction struct {
	SendMsg        []byte
	SequenceNumber uint32
	MessageType    MessageType
	TxType         TransactionType
	EventChannel   chan EventType
	Conn           *net.UDPConn
	DestAddr       *net.UDPAddr
	ConsumerAddr   string
}

// NewTransaction - create pfcp transaction object
func NewTransaction(pfcpMSG Message, binaryMSG []byte, Conn *net.UDPConn, DestAddr *net.UDPAddr) (tx *Transaction) {
	tx = &Transaction{
		SendMsg:        binaryMSG,
		SequenceNumber: pfcpMSG.Header.SequenceNumber,
		MessageType:    pfcpMSG.Header.MessageType,
		EventChannel:   make(chan EventType),
		Conn:           Conn,
		DestAddr:       DestAddr,
	}

	if pfcpMSG.IsRequest() {
		tx.TxType = SendingRequest
		tx.ConsumerAddr = Conn.LocalAddr().String()
	} else if pfcpMSG.IsResponse() {
		tx.TxType = SendingResponse
		tx.ConsumerAddr = DestAddr.String()
	}

	logger.PFCPLog.Tracef("New Transaction SEQ[%d] DestAddr[%s]", tx.SequenceNumber, DestAddr.String())
	return
}

func (transaction *Transaction) Start() {

	logger.PFCPLog.Tracef("Start Transaction [%d]\n", transaction.SequenceNumber)

	if transaction.TxType == SendingRequest {
		for iter := 0; iter < NumOfResend; iter++ {
			timer := time.NewTimer(ResendRequestTimeOutPeriod * time.Second)

			_, err := transaction.Conn.WriteToUDP(transaction.SendMsg, transaction.DestAddr)

			if err != nil {
				logger.PFCPLog.Warnf("Request Transaction [%d]: %s\n", transaction.SequenceNumber, err)
				return
			}

			select {
			case event := <-transaction.EventChannel:

				if event == ReceiveValidResponse {
					logger.PFCPLog.Tracef("Request Transaction [%d]: receive valid response\n", transaction.SequenceNumber)
					return
				}
			case <-timer.C:
				logger.PFCPLog.Tracef("Request Transaction [%d]: timeout expire\n", transaction.SequenceNumber)
				logger.PFCPLog.Tracef("Request Transaction [%d]: Resend packet\n", transaction.SequenceNumber)
				continue
			}
		}
	} else if transaction.TxType == SendingResponse {
		//Todo :Implement SendingResponse type of reliable delivery
		timer := time.NewTimer(ResendResponseTimeOutPeriod * time.Second)
		for iter := 0; iter < NumOfResend; iter++ {

			_, err := transaction.Conn.WriteToUDP(transaction.SendMsg, transaction.DestAddr)

			if err != nil {
				logger.PFCPLog.Warnf("Response Transaction [%d]: sending error\n", transaction.SequenceNumber)
				return
			}

			select {
			case event := <-transaction.EventChannel:

				if event == ReceiveResendRequest {
					logger.PFCPLog.Tracef("Response Transaction [%d]: receive resend request\n", transaction.SequenceNumber)
					logger.PFCPLog.Tracef("Response Transaction [%d]: Resend packet\n", transaction.SequenceNumber)
					continue
				}
			case <-timer.C:
				logger.PFCPLog.Tracef("Response Transaction [%d]: timeout expire\n", transaction.SequenceNumber)
				return
			}
		}

	}

}
