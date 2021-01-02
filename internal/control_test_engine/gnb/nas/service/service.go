package service

import (
	"log"
	"my5G-RANTester/internal/control_test_engine/gnb/context"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func initServer(gnb *context.GNBContext) {

	// initiated communication GNB server with unix sockets.
	log.Println("Starting Unix server")
	ln, err := net.Listen("unix", "/tmp/gnb.sock")
	if err != nil {
		log.Fatal("Listen error: ", err)
	}

	gnb.SetListener(ln)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func(ln net.Listener, c chan os.Signal) {
		sig := <-c
		log.Printf("Caught signal %s: shutting down.", sig)
		ln.Close()
		os.Exit(0)
	}(ln, sigc)
}
