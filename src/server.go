package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// RawPacket is the raw-bytes received over wire, with the RemoteAddr saved
type RawPacket struct {
	Addr  *net.UDPAddr
	bytes []byte
}

func (packet RawPacket) getBytes() []byte {
	return packet.bytes
}

func (packet RawPacket) getAddr() *net.UDPAddr {
	return packet.Addr
}

// ListenAndServe is the engine for the tftp-server
func ListenAndServe(serverIPPort string, numThreads int, timeout int) {

	addr, err := net.ResolveUDPAddr("udp", serverIPPort)
	if err != nil {
		fmt.Printf("ERROR: ResolveUDPAddr()::serverIPPort:[%s]\n", serverIPPort)
		os.Exit(1)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Printf("ERROR: ListenUDP()::addr:[%s]\n", addr)
		os.Exit(2)
	}

	fmt.Println("Listening: " + serverIPPort)

	// Create a channel of size = # of threads
	dataChannel := make(chan RawPacket, numThreads)
	defer close(dataChannel)

	fmt.Printf("Threads: %d Starting...", numThreads)
	for i := 0; i < numThreads; i++ {
		go processProtocol(dataChannel, timeout)
	}
	fmt.Printf("Done\n")

	fmt.Printf("Listener: Running\n")
	for {
		buf := make([]byte, MaxPacketSize)
		n, remoteAddr, _ := conn.ReadFromUDP(buf)

		fmt.Printf("Packet::n:[%d]\n", n)
		fmt.Printf("Packet::remoteAddr:[%s]\n", remoteAddr)

		request := RawPacket{
			Addr:  remoteAddr,
			bytes: buf[:n],
		}

		dataChannel <- request
	}

}

func processProtocol(dataChannel chan RawPacket, timeout int) {

	nexus := NewFileNexus()

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

		opcode, p, err := ParsePacket(rawRequestBuffer)
		switch opcode {
		case OpRRQ:
			packetReq := makePacketRequest(p.Serialize())
			doReadReq(nexus, conn, rawPacket.Addr, packetReq)
		case OpWRQ:
			print("OpWRQ\n")
			//fmt.Println("Write... type:", reflect.TypeOf(p).Elem(), " packet:", p)
		case OpData:
			print("OpData\n")
		case OpAck:
			print("OpAck\n")
		case OpError:
			print("OpError\n")
		}

		fmt.Printf("*** conn.Close() ***\n")
		conn.Close()
	}
}

func doSendError(conn *net.UDPConn, code uint16, msg string) {
	fmt.Printf("doSendError()::msg:[%s]\n", msg)
	p := NewPacketError(code, msg)
	conn.Write(p.Serialize())
}

func doReadReq(nexus *FileNexus, conn *net.UDPConn, remoteAddr *net.UDPAddr, packet PacketRequest) {

	fmt.Printf("doReadReq()::remoteAddr():[%s]\n", remoteAddr.String())

	// Validate OpMode
	if strings.Compare(strings.ToLower(packet.Mode), "octect") == 0 {
		errmsg := fmt.Sprintf("ERROR: mode:[%s] is not supported.\n", packet.Mode)
		fmt.Printf("%s\n", errmsg)
		doSendError(conn, ErrorNotDefined, errmsg)
		conn.Close()
		return
	}

	// Load the File into Nexus
	ok, entry := nexus.GetEntry(conn, packet.Filename)
	if !ok {
		return
	}

	var curBlock uint16 = 1
	var curPos int = 0

	// Create ACK Packet (Reusable)
	ackPacket := PacketAck{}
	ackBuffer := make([]byte, 4) // ?? sizeof(PacketAck)

	for curPos < len(entry.Bytes) {

		// Set the PacketSize with bounds to the end of file
		packetSize := curPos + MaxDataBlockSize
		if packetSize > len(entry.Bytes) {
			packetSize = len(entry.Bytes)
		}

		// Send the Data Packet
		dataPacket := makePacketData(curBlock, entry.Bytes)
		_, err := conn.WriteToUDP(dataPacket.Serialize(), remoteAddr)
		if err != nil {
			errmsg := fmt.Sprintf("ERROR:[%s] doReadReq()::conn.WriteToUDP()::remoteAddr:[%s]", err.Error(), remoteAddr.String())
			doSendError(conn, ErrorNotDefined, errmsg)
			conn.Close() // Should I really do this?!?
			return
		}

		// ACK PACKET!
		for {
			_, readRemoteAddr, err := conn.ReadFromUDP(ackBuffer)
			if err != nil {
				errmsg := fmt.Sprintf("ERROR:[%s] doReadReq()::conn.Read()::readRemoteAdrr:[%s]\n", err.Error(), readRemoteAddr)
				doSendError(conn, ErrorNotDefined, errmsg)
				// conn.Close() // Should I really do this?!?
				return
			}

			if readRemoteAddr.Port != remoteAddr.Port {
				// Packet from unknown host
				errmsg := fmt.Sprintf("ERROR: doReadReq()::remoteAddr.Port:[%d] != readRemoteAddr.Port:[%d] ", remoteAddr.Port, readRemoteAddr.Port)
				doSendError(conn, ErrorUnknownTID, errmsg)
				// conn.Close() // Should I really do this?!?
				continue
			}
			break
		}

		err = ackPacket.Parse(ackBuffer)
		if err != nil {
			errmsg := fmt.Sprintf("ERROR:[%s] doReadReq()::AckPacket.Parse()", err.Error())
			doSendError(conn, ErrorNotDefined, errmsg) // ?? @TODO Is this an OP error?
			return
		}

		curPos = curPos + packetSize
		curBlock = ackPacket.BlockNum + 1
	}
}
