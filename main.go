package main

import "gateServer/gate"

func main() {
	gate := &gate.Gate{
		MaxConnNum:   20000,
		MaxMsgLen:    4096,
		TCPAddr:      "127.0.0.1:3563",
		LenMsgLen:    2,
		LittleEndian: false,
	}
	closeSig := make(chan bool, 1)
	gate.Run(closeSig)
}
