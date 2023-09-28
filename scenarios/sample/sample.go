package main

import "fmt"

//export attach
func attach(uint32)
//export detach
func detach(uint32)
//export pduSessionRequest
func pduSessionRequest(uint32, uint32)
//export pduSessionRelease
func pduSessionRelease(uint32, uint32)
//export think
func think(uint32)

//export ueHandler
func ueHandler(ueId uint32)  {
	fmt.Println("Hi, I'm an UE that wants to attach!")
	attach(ueId)

	think(5000)

	fmt.Println("Hi, I'm an UE that wants to have 16 PDU Sessions!")
	for pduSessionId:=uint32(1); pduSessionId<2; pduSessionId++ {
		fmt.Println("Hi, I'm an UE that wants to request PDU Session id: ", pduSessionId)
		pduSessionRequest(ueId, pduSessionId)
	}

	think(5000)

	fmt.Println("Hi, I'm an UE that wants to release 16 PDU Sessions!")
	for pduSessionId:=uint32(1); pduSessionId<2; pduSessionId++ {
		fmt.Println("Hi, I'm an UE that wants to release PDU Session id: ", pduSessionId)
		pduSessionRelease(ueId, pduSessionId)
	}

	fmt.Println("Hi, I'm an UE that wants to detach!")
	detach(ueId)
}


func main() {}
