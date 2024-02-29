/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 */
package sender

import (
	"fmt"
	"my5G-RANTester/lib/ngap/ngapSctp"
	"time"

	"github.com/ishidawataru/sctp"
)

var senderInstance Sender

type Sender interface {
	sendToAmF([]byte, *sctp.SCTPConn) error
}

func SendToAmF(message []byte, conn *sctp.SCTPConn) error {
	if senderInstance == nil {
		return fmt.Errorf("Error sending NGAP message: sender not initialized")
	}
	return senderInstance.sendToAmF(message, conn)
}

type defaultSender struct{}

func (s *defaultSender) sendToAmF(message []byte, conn *sctp.SCTPConn) error {
	return send(message, conn)
}

type senderWithRate struct {
	MaxDebit int // requests per seconds
	Queue    chan sendctx
}

type sendctx struct {
	message []byte
	conn    *sctp.SCTPConn
}

func (s *senderWithRate) sendToAmF(message []byte, conn *sctp.SCTPConn) error {
	s.Queue <- sendctx{message, conn}
	return nil
}

func (s *senderWithRate) start() {
	val := int(time.Second.Abs()) / s.MaxDebit
	ticker := time.NewTicker(time.Duration(val))
	go func() {
		for {
			<-ticker.C
			ctx := <-s.Queue
			send(ctx.message, ctx.conn)
		}
	}()
}

func Init(debit int) {
	if debit > 0 {
		sender := senderWithRate{
			MaxDebit: debit,
			Queue:    make(chan sendctx, 50),
		}
		sender.start()
		senderInstance = &sender
	} else {
		senderInstance = &defaultSender{}
	}
}

func send(message []byte, conn *sctp.SCTPConn) error {
	// TODO included information for SCTP association.
	info := &sctp.SndRcvInfo{
		Stream: uint16(0),
		PPID:   ngapSctp.NGAP_PPID,
	}

	_, err := conn.SCTPWrite(message, info)
	if err != nil {
		return fmt.Errorf("Error sending NGAP message %q", err)
	}

	return nil
}
