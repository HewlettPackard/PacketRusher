package pfcpUdp

import (
	"fmt"
	"free5gc/lib/pfcp"
	"free5gc/lib/pfcp/logger"
	"net"
)

const (
	PFCP_PORT        = 8805
	PFCP_MAX_UDP_LEN = 2048
)

type PfcpServer struct {
	Addr string
	Conn *net.UDPConn
	//Consumer Table
	//Map Consumer IP to its tx table
	ConsumerTable map[string]pfcp.TxTable
}

func NewPfcpServer(addr string) (server PfcpServer) {
	server.Addr = addr
	server.ConsumerTable = make(map[string]pfcp.TxTable)

	return
}

func (pfcpServer *PfcpServer) Listen() error {
	var serverIp net.IP
	if pfcpServer.Addr == "" {
		serverIp = net.IPv4zero
	} else {
		serverIp = net.ParseIP(pfcpServer.Addr)
	}

	addr := &net.UDPAddr{
		IP:   serverIp,
		Port: PFCP_PORT,
	}

	conn, err := net.ListenUDP("udp", addr)
	pfcpServer.Conn = conn
	return err
}

func (pfcpServer *PfcpServer) ReadFrom(msg *pfcp.Message) (*net.UDPAddr, error) {
	buf := make([]byte, PFCP_MAX_UDP_LEN)
	n, addr, err := pfcpServer.Conn.ReadFromUDP(buf)
	if err != nil {
		return addr, err
	}

	err = msg.Unmarshal(buf[:n])
	if err != nil {
		return addr, err
	}

	if msg.IsRequest() {
		//Todo: Implement SendingResponse type of reliable delivery
		tx, err := pfcpServer.FindTransaction(msg, addr)

		if err != nil {
			return addr, err
		}

		if tx != nil {
			err = fmt.Errorf("Receive resend PFCP request")
			tx.EventChannel <- pfcp.ReceiveResendRequest
			return addr, err
		}

	} else if msg.IsResponse() {
		tx, err := pfcpServer.FindTransaction(msg, pfcpServer.Conn.LocalAddr().(*net.UDPAddr))
		if err != nil {
			return addr, err
		}

		tx.EventChannel <- pfcp.ReceiveValidResponse
	}

	return addr, nil
}

func (pfcpServer *PfcpServer) WriteTo(msg pfcp.Message, addr *net.UDPAddr) error {
	buf, err := msg.Marshal()
	if err != nil {
		return err
	}

	/*TODO: check if all bytes of buf are sent*/
	tx := pfcp.NewTransaction(msg, buf, pfcpServer.Conn, addr)

	err = pfcpServer.PutTransaction(tx)
	if err != nil {
		return err
	}

	go pfcpServer.StartTxLifeCycle(tx)
	return nil
}

func (pfcpServer *PfcpServer) Close() error {
	return pfcpServer.Conn.Close()
}

func (pfcpServer *PfcpServer) PutTransaction(tx *pfcp.Transaction) (err error) {

	logger.PFCPLog.Traceln("In PutTransaction")

	consumerAddr := tx.ConsumerAddr
	if _, exist := pfcpServer.ConsumerTable[consumerAddr]; !exist {

		pfcpServer.ConsumerTable[consumerAddr] = make(pfcp.TxTable)
	}

	txTable := pfcpServer.ConsumerTable[consumerAddr]
	if _, exist := txTable[tx.SequenceNumber]; !exist {

		txTable[tx.SequenceNumber] = tx
	} else {

		logger.PFCPLog.Warnln("In PutTransaction")
		logger.PFCPLog.Warnln("Consumer Addr: ", consumerAddr)
		logger.PFCPLog.Warnln("Sequence number ", tx.SequenceNumber, " already exist!")
		err = fmt.Errorf("Insert tx error: duplicate sequence number %d", tx.SequenceNumber)
	}

	logger.PFCPLog.Traceln("End PutTransaction")
	return
}

func (pfcpServer *PfcpServer) RemoveTransaction(tx *pfcp.Transaction) (err error) {

	logger.PFCPLog.Traceln("In RemoveTransaction")
	consumerAddr := tx.ConsumerAddr
	txTable := pfcpServer.ConsumerTable[consumerAddr]

	if tx, exist := txTable[tx.SequenceNumber]; exist {

		if tx.TxType == pfcp.SendingRequest {
			logger.PFCPLog.Infof("Remove Request Transaction [%d]\n", tx.SequenceNumber)
		} else if tx.TxType == pfcp.SendingResponse {
			logger.PFCPLog.Infof("Remove Request Transaction [%d]\n", tx.SequenceNumber)
		}

		delete(txTable, tx.SequenceNumber)
	} else {

		logger.PFCPLog.Warnln("In RemoveTransaction")
		logger.PFCPLog.Warnln("Consumer IP: ", consumerAddr)
		logger.PFCPLog.Warnln("Sequence number ", tx.SequenceNumber, " doesn't exist!")
		err = fmt.Errorf("Remove tx error: transaction [%d] doesn't exist\n", tx.SequenceNumber)
	}

	logger.PFCPLog.Traceln("End RemoveTransaction")
	return
}

func (pfcpServer *PfcpServer) StartTxLifeCycle(tx *pfcp.Transaction) {
	//Start Transaction
	tx.Start()

	//End Transaction
	err := pfcpServer.RemoveTransaction(tx)
	if err != nil {
		logger.PFCPLog.Warnln(err)
	}
}

func (pfcpServer *PfcpServer) FindTransaction(msg *pfcp.Message, addr *net.UDPAddr) (tx *pfcp.Transaction, err error) {

	logger.PFCPLog.Traceln("In FindTransaction")
	consumerAddr := addr.String()

	if msg.IsResponse() {
		if _, exist := pfcpServer.ConsumerTable[consumerAddr]; !exist {
			logger.PFCPLog.Warnln("In FindTransaction")
			logger.PFCPLog.Warnf("Can't find txTable from consumer addr: [%s]", consumerAddr)
			err = fmt.Errorf("FindTransaction Error: txTable not found")
			return
		}

		txTable := pfcpServer.ConsumerTable[consumerAddr]
		seqNum := msg.Header.SequenceNumber

		if _, exist := txTable[seqNum]; !exist {
			logger.PFCPLog.Warnln("In FindTransaction")
			logger.PFCPLog.Warnln("Consumer Addr: ", consumerAddr)
			logger.PFCPLog.Warnf("Can't find tx [%d] from txTable: ", seqNum)
			err = fmt.Errorf("FindTransaction Error: sequence number [%d] not found", seqNum)
			return
		}

		tx = txTable[seqNum]
	} else if msg.IsRequest() {
		if _, exist := pfcpServer.ConsumerTable[consumerAddr]; !exist {
			return
		}

		txTable := pfcpServer.ConsumerTable[consumerAddr]
		seqNum := msg.Header.SequenceNumber

		if _, exist := txTable[seqNum]; !exist {
			return
		}

		tx = txTable[seqNum]
	}
	logger.PFCPLog.Traceln("End FindTransaction")
	return

}

// Send a PFCP message and close UDP connection
func SendPfcpMessage(msg pfcp.Message, srcAddr *net.UDPAddr, dstAddr *net.UDPAddr) error {
	conn, err := net.DialUDP("udp", srcAddr, dstAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	buf, err := msg.Marshal()
	if err != nil {
		return err
	}

	/*TODO: check if all bytes of buf are sent*/
	_, err = conn.Write(buf)
	if err != nil {
		return err
	}

	return nil
}

// Receive a PFCP message and close UDP connection
func ReceivePfcpMessage(msg *pfcp.Message, srcAddr *net.UDPAddr, dstAddr *net.UDPAddr) error {
	conn, err := net.DialUDP("udp", srcAddr, dstAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	buf := make([]byte, PFCP_MAX_UDP_LEN)
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}

	err = msg.Unmarshal(buf[:n])
	if err != nil {
		return err
	}

	return nil
}
