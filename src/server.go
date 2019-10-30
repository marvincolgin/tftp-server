package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// SetupListener will establish a listener on the given Server IP/Port
func SetupListener(serverIPPort string) *net.UDPConn {

	addr, err := net.ResolveUDPAddr("udp", serverIPPort)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: ResolveUDPAddr()::serverIPPort:[%s]\n", serverIPPort)
		os.Exit(1)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: ListenUDP()::addr:[%s]\n", addr)
		os.Exit(2)
	}

	fmt.Fprintln(os.Stdout, "Listening: "+serverIPPort)

	return conn
}

// ListenAndServe is the engine for the tftp-server
func ListenAndServe(serverIPPort string, numThreads int, timeout int) {

	// Listener Start
	conn := SetupListener(serverIPPort)

	// Create a *buffered* channel = # of threads, as one thread per channel to prevent blocks/dropped data
	dataChannel := make(chan RawPacket, numThreads)
	defer close(dataChannel)

	// Central repo for File data and mutexes
	nexus := NewFileNexus()

	// Create threads and pass the dataChannel
	fmt.Fprintf(os.Stdout, "Threads: %d Starting...", numThreads)
	for i := 0; i < numThreads; i++ {
		go processProtocol(nexus, dataChannel, timeout)
	}
	fmt.Fprintf(os.Stdout, "Done\n")

	// Forever Loop...Listening
	fmt.Fprintf(os.Stdout, "Listener: Loop Running\n")
	rcvBuf := make([]byte, MaxPacketSize)
	for {

		// Blocking read from Listener
		cnt, remoteAddr, _ := conn.ReadFromUDP(rcvBuf)

		// Bundle raw packet bytes with IP, as thread won't have access to "conn"
		rawPacket := RawPacket{
			Addr:  remoteAddr,
			bytes: rcvBuf[:cnt],
		}

		// fan-out the bytes to the *buffered* channel for goroutines to process
		dataChannel <- rawPacket
	}

}

func createUDPEndPoint(addr string, port int) (bool, *net.UDPAddr, *net.UDPConn) {

	// Establish Connection
	localAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: resolveUDPAddr(). Dumping Packet")
		return false, nil, nil
	}
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: listenUDP(). Dumping Packet")
		return false, localAddr, nil
	}

	return true, localAddr, conn
}

// processProtocol goroutine to process data received by main-thread and "fanned out"
func processProtocol(nexus *FileNexus, dataChannel chan RawPacket, timeout int) {

	for {

		// read packet out of the channel to process
		rawPacket := <-dataChannel

		success, _, conn := createUDPEndPoint("", 0)
		if !success {
			continue
		}

		// get raw bytes from packet
		rawRequestBuffer := rawPacket.getBytes()

		opcode, p, _ := ParsePacket(rawRequestBuffer) // @TODO discarded err
		switch opcode {
		case OpRRQ:
			// @TODO re-evaluate this..., do I need makePacketRequest, can I use wire.go?
			packetReq := makePacketRequest(p.Serialize())
			doReadReq(nexus, conn, rawPacket.Addr, packetReq)
		case OpWRQ:
			fmt.Fprintf(os.Stdout, "OpWRQ\n")
			//fmt.Println("Write... type:", reflect.TypeOf(p).Elem(), " packet:", p)
		case OpData:
			fmt.Fprintf(os.Stdout, "OpData\n")
		case OpAck:
			fmt.Fprintf(os.Stdout, "OpAck\n")
		case OpError:
			fmt.Fprintf(os.Stdout, "OpError\n")
		}

		// Close the connection as we are done processing the packet
		conn.Close()
	}
}

func doSendError(conn *net.UDPConn, code uint16, msg string) {
	fmt.Fprintf(os.Stderr, "doSendError()::msg:[%s]\n", msg)
	p := NewPacketError(code, msg)
	conn.Write(p.Serialize())
}

func doReadReq(nexus *FileNexus, conn *net.UDPConn, remoteAddr *net.UDPAddr, packet PacketRequest) {

	fmt.Fprintf(os.Stdout, "doReadReq()::remoteAddr():[%s]\n", remoteAddr.String())

	// Validate OpMode
	if strings.Compare(strings.ToLower(packet.Mode), "octect") == 0 {
		errmsg := fmt.Sprintf("ERROR: mode:[%s] is not supported.\n", packet.Mode)
		fmt.Fprintln(os.Stderr, errmsg)
		doSendError(conn, ErrorNotDefined, errmsg)
		conn.Close()
		return
	}

	// Load the File into Nexus
	ok, entry := nexus.GetEntry(conn, packet.Filename)
	if !ok {
		return
	}

	// Create ACK Packet (Reusable)
	ackPacket := PacketAck{}
	ackBuffer := make([]byte, 4) // ?? sizeof(PacketAck)

	// Loop through the entire file
	var curBlock uint16 = 1
	var curPos int = 0

	for curPos <= len(entry.Bytes) {

		// *** @TODO ZERO BYTE WILL SIGNAL END

		// Set the PacketSize with bounds to the end of file
		packetSize := MaxDataBlockSize
		if curPos+packetSize > len(entry.Bytes) {
			packetSize = len(entry.Bytes) - curPos
		}

		// fmt.Fprintln(os.Stdout, "STATUS: curPos:[%d] curBlock:[%d] len(entry.Bytes):[%d] packetSize:[%d]\n", curPos, curBlock, len(entry.Bytes), packetSize)

		// Send the Data Packet
		dataPacket := makePacketData(curBlock, entry.Bytes, curPos, packetSize)
		_, err := conn.WriteToUDP(dataPacket.Serialize(), remoteAddr)
		if err != nil {
			errmsg := fmt.Sprintf("ERROR:[%s] doReadReq()::conn.WriteToUDP()::remoteAddr:[%s]", err.Error(), remoteAddr.String())
			doSendError(conn, ErrorNotDefined, errmsg)
			conn.Close() // @TODO: Should I really do this?!?
			return
		}

		// WAIT for our ACK packet
		for {
			_, readRemoteAddr, err := conn.ReadFromUDP(ackBuffer)
			if err != nil {
				errmsg := fmt.Sprintf("ERROR:[%s] doReadReq()::conn.Read()::readRemoteAdrr:[%s]\n", err.Error(), readRemoteAddr)
				doSendError(conn, ErrorNotDefined, errmsg)
				// conn.Close() I should *NOT* do this, as it's an error packet, wait for retry
				return
			}

			if readRemoteAddr.Port != remoteAddr.Port {
				// Packet from unknown host
				errmsg := fmt.Sprintf("ERROR: doReadReq()::remoteAddr.Port:[%d] != readRemoteAddr.Port:[%d] ", remoteAddr.Port, readRemoteAddr.Port)
				doSendError(conn, ErrorUnknownTID, errmsg)
				// conn.Close() I should *NOT* do this, it's not a reason to disconnect, it's just a bogus packet
				continue
			}

			// We got our ACK, if we are here...
			break
		}

		// Parse the ACK Packet
		err = ackPacket.Parse(ackBuffer)
		if err != nil {
			errmsg := fmt.Sprintf("ERROR:[%s] doReadReq()::AckPacket.Parse()", err.Error())
			doSendError(conn, ErrorNotDefined, errmsg) // ?? @TODO Is this an OP error?
			return
		}
		// Set current block to be the ackPacket's blocknum (as it could have incremented this value in resends of Ack)
		curBlock = ackPacket.BlockNum + 1

		// Advance our position in the file
		if packetSize == 0 {
			// this last packet was a terminating packet, as it's is 0 bytes and it just so happens
			break
		} else {
			curPos = curPos + packetSize
		}

	}

	fmt.Fprintf(os.Stdout, "SUCCESS: transferred file:[%s] to client:[%s]\n", packet.Filename, remoteAddr.String())
}
