package main

import (
	"fmt"
	"net"
	"os"
)

// RawPacket is the raw-bytes received over wire, with the RemoteAddr saved
type RawPacket struct {
	Addr   *net.UDPAddr
	buffer []byte
}

func (packet RawPacket) getBytes() []byte {
	return packet.buffer
}

func (packet RawPacket) getAddr() *net.UDPAddr {
	return packet.Addr
}

func abortError(err error) {
	if err != nil {
		fmt.Println("ERROR: " + err.Error())
		os.Exit(1)
	}
}

func launch(serverIPPort string, numThreads int, timeout int) {

	addr, err := net.ResolveUDPAddr("udp", serverIPPort)
	abortError(err)

	conn, err := net.ListenUDP("udp", addr)
	abortError(err)

	fmt.Println("Launched: " + serverIPPort)

	// Create a channel of size = # of threads
	dataChannel := make(chan RawPacket, numThreads)

	for i := 0; i < numThreads; i++ {
		go processProtocol(dataChannel, timeout)
	}

	for {
		buffer := make([]byte, MaxPacketSize)
		n, remoteAddr, _ := conn.ReadFromUDP(buffer)

		request := RawPacket{
			Addr:   remoteAddr,
			buffer: buffer[:n],
		}

		dataChannel <- request
	}

	close(dataChannel)
}

func processProtocol(dataChannel chan RawPacket, timeout int) {

	for {
		rawPacket := <-dataChannel

		// Establish Connection
		localAddr, err := net.ResolveUDPAddr("udp", ":0")
		if err != nil {
			fmt.Println("ERROR: resolveUDPAddr(). Dumping Packet")
			continue
		}
		conn, err := net.ListenUDP("udp", localAddr)
		if err != nil {
			fmt.Println("ERROR: listenUDP(). Dumping Packet")
			continue
		}

		// get raw bytes from packet
		rawRequestBuffer := rawPacket.getBytes()

		opcode, _, err := ParsePacket(rawRequestBuffer)
		switch opcode {
		case OpRRQ:
			print("OpRRQ\n")
		case OpWRQ:
			print("OpWRQ\n")
		case OpData:
			print("OpData\n")
		case OpAck:
			print("OpAck\n")
		case OpError:
			print("OpError\n")
		}

		conn.Close()
	}
}
