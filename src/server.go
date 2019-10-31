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
		fmt.Fprintf(os.Stderr, "ERROR: ResolveUDPAddr()::serverIPPort:[%s]::err.Error():[%s]\n", serverIPPort, err.Error())
		os.Exit(1)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: ListenUDP()::addr:[%s]::err.Error():[%s]\n", addr, err.Error())
		os.Exit(2)
	}

	fmt.Fprintf(os.Stdout, "Listener: %s\n", serverIPPort)

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
		fmt.Fprintf(os.Stderr, "ERROR: resolveUDPAddr()::err.Error():[%s]\n", err.Error())
		return false, nil, nil
	}
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: listenUDP()::err.Error():[%s]\n", err.Error())
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

		opcode, p, err := ParsePacket(rawRequestBuffer) // @TODO discarded err
		if err == nil {
			switch opcode {
			case OpRRQ:
				// @TODO re-evaluate this..., do I need makePacketRequest, can I use wire.go?
				packetReq := makePacketRequest(p.Serialize())
				doReadReq(nexus, conn, rawPacket.Addr, packetReq)
			case OpWRQ:
				// @TODO re-evaluate this..., do I need makePacketRequest, can I use wire.go?
				packetReq := makePacketRequest(p.Serialize())
				doWriteReq(nexus, conn, rawPacket.Addr, packetReq)
			default:
				fmt.Fprintf(os.Stderr, "processProtocol()::Invalid Opcode::opcode:[%d]", opcode)
			}
		} else {
			fmt.Fprintf(os.Stderr, "processProtocol()::ParsePacket()::err.Error():[%s]\n", err.Error())
		}

		// Close the connection as we are done processing the packet
		conn.Close()
	}
}

// doSendError will send an error packet on conn to client
func doSendError(conn *net.UDPConn, code uint16, msg string) {
	fmt.Fprintf(os.Stderr, "doSendError()::msg:[%s]\n", msg)
	p := NewPacketError(code, msg)
	conn.Write(p.Serialize())
}

// doValidateOpMode we only support binary aka octect at this time
func doValidateOpMode(conn *net.UDPConn, mode string) bool {

	if strings.Compare(strings.ToLower(mode), "octect") == 0 {
		errmsg := fmt.Sprintf("ERROR: mode:[%s] is not supported.\n", mode)
		fmt.Fprintln(os.Stderr, errmsg)
		doSendError(conn, ErrorNotDefined, errmsg)
		conn.Close()
		return false
	}
	return true
}

// doReadReq will process the incoming request packet and continue until file req processed
func doReadReq(nexus *FileNexus, conn *net.UDPConn, remoteAddr *net.UDPAddr, packet PacketRequest) {

	fmt.Fprintf(os.Stdout, "READ: REQUEST file:[%s], client:[%s]\n", packet.Filename, remoteAddr.String())

	// Validate OpMode
	if !doValidateOpMode(conn, packet.Mode) {
		return
	}

	// Load the File into Nexus
	ok, entry := nexus.GetEntry(conn, remoteAddr.String(), packet.Filename)
	if !ok {
		return
	}

	// File Exists?
	if entry.Bytes == nil {
		errmsg := fmt.Sprintf("ERROR: Requested file does not exist, file:[%s]", packet.Filename)
		doSendError(conn, ErrorFileNotFound, errmsg)
		conn.Close()
		return
	}

	// Create ACK Packet (Reusable)
	ackPacket := PacketAck{}
	ackBuffer := make([]byte, 4)

	// Loop through the entire file
	var curBlock uint16 = 1
	var curPos int = 0

	for curPos <= len(entry.Bytes) {

		// Set the PacketSize with bounds to the end of file
		packetSize := MaxDataBlockSize
		if curPos+packetSize > len(entry.Bytes) {
			packetSize = len(entry.Bytes) - curPos
		}

		// fmt.Fprintf(os.Stdout, "READ: STATUS curPos:[%d] curBlock:[%d] packetSize:[%d] len(entry.Bytes):[%d]\n", curPos, curBlock, packetSize, len(entry.Bytes))

		// Send the Data Packet
		dataPacket := makePacketData(curBlock, entry.Bytes, curPos, packetSize)
		_, err := conn.WriteToUDP(dataPacket.Serialize(), remoteAddr)
		if err != nil {
			errmsg := fmt.Sprintf("ERROR:[%s] doReadReq()::conn.WriteToUDP()::remoteAddr:[%s]", err.Error(), remoteAddr.String())
			doSendError(conn, ErrorNotDefined, errmsg)
			conn.Close() // @TODO: Should I really do this?!?
			return
		}

		// End of the Line! We just sent a ZERO byte packet, so that's the end of the transfer and we exit
		// NOTE: We did this, cuz "for curPos <= len(entry.Bytes)", which got us here
		if curPos == len(entry.Bytes) {
			break
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

	fmt.Fprintf(os.Stdout, "READ: SUCCESS file:[%s], client:[%s]\n", packet.Filename, remoteAddr.String())
}

// doWriteReq will process the incoming request packet and continue until file req processed
func doWriteReq(nexus *FileNexus, conn *net.UDPConn, remoteAddr *net.UDPAddr, packet PacketRequest) {

	fmt.Fprintf(os.Stdout, "WRITE: REQUEST file:[%s], client:[%s]\n", packet.Filename, remoteAddr.String())

	// Validate OpMode
	if !doValidateOpMode(conn, packet.Mode) {
		return
	}

	// Load the File into Nexus
	ok, entry := nexus.GetEntry(conn, remoteAddr.String(), packet.Filename)
	if !ok {
		return
	}

	// Zero out the file
	if len(entry.Bytes) > 0 {
		entry.Bytes = nil
	}

	// Create ACK Packet (Reusable)
	ackPacket := PacketAck{}
	packetData := PacketData{}
	rcvBuf := make([]byte, MaxPacketSize) // Data comes in as 2048 packets

	// First Block will be zero (0) in response to REQ
	var curBlock uint16 = 0

	cntReadActual := MaxDataBlockSize + 4 // 4 bytes is BlockNum?

	for {

		// Send the ACK Packet
		ackPacket.BlockNum = curBlock
		_, err := conn.WriteToUDP(ackPacket.Serialize(), remoteAddr)
		if err != nil {
			errmsg := fmt.Sprintf("ERROR:[%s] doWriteReq()::conn.WriteToUDP()::remoteAddr:[%s]", err.Error(), remoteAddr.String())
			doSendError(conn, ErrorNotDefined, errmsg)
			return
		}

		// Short Packet, most likely the END
		if cntReadActual < MaxDataBlockSize+4 {

			// 4 Bytes is the official END (as it's a data packet with no DATA).. otherwise, it is an error
			if cntReadActual != 4 {
				// I thought this was an issue, but it doesn't appear to be.. wasn't sure these last bytes were getting there
				// fmt.Fprintf(os.Stderr, "DEBUG: doWriteReq()::cntReadActual:[%d]\n", cntReadActual)
			} else {
				// Normal
				// fmt.Fprintf(os.Stdout, "WRITE: Got Short Packet, 4 bytes.\n")
			}
			break
		}

		var cntReadFromUDP int = 0
		var clientAddr *net.UDPAddr

		for {

			// READ off UDP connection
			cntReadFromUDP, clientAddr, err = conn.ReadFromUDP(rcvBuf)
			if err != nil {
				errmsg := fmt.Sprintf("ERROR: doWriteReq()::conn.ReadFromUDP()::remoteAddr:[%s]::err.Error():[%s]\n", remoteAddr, err.Error())
				doSendError(conn, ErrorNotDefined, errmsg)
				return
			}

			if clientAddr.Port != remoteAddr.Port {
				errmsg := fmt.Sprintf("ERROR: doWriteReq()::clientAddr.Port!=remoteAddr.Port::clientAddr.Port:[%d]::remoteAddr.Port:[%d]\n", clientAddr.Port, remoteAddr.Port)
				doSendError(conn, ErrorUnknownTID, errmsg)
				continue
			}
			break
		}

		err = packetData.Parse(rcvBuf)
		if err != nil {
			errmsg := fmt.Sprintf("ERROR: doWriteReq()::PacketData.Parse()::err.Error():[%s]", err.Error())
			doSendError(conn, ErrorIllegalOp, errmsg)
			return
		}

		// fmt.Fprintf(os.Stdout, "WRITE: STATUS curBlock:[%d] len(entry.Bytes):[%d] cntReadFromUDP:[%d]\n", curBlock, len(entry.Bytes), cntReadFromUDP)

		// Out of order, as this isn't the next seq block req. As a result, we will loop and re-ack what we want
		if packetData.BlockNum-1 != curBlock {
			continue
		}

		// Append the new Bytes to the end
		// @TODO optimize: make Nexus func to perform this work, but alloc an ever increasing size and maintain a
		// 		length variable of data used in alloc (this will prevent the thrashing of memory to constantly move
		//      this array around to seq memory)
		if cntReadFromUDP > 4 {
			entry.Bytes = append(entry.Bytes, packetData.Data[:cntReadFromUDP-4]...) // NOTE: Slice is used: 4 bytes for OP&BlockNum, then the rest of the data
		}
		cntReadActual = cntReadFromUDP
		curBlock = curBlock + 1

	}

	// COMPLETE: Output and Save File
	fmt.Fprintf(os.Stdout, "WRITE: SUCCESS file:[%s], bytes:[%d], client:[%s]\n", packet.Filename, len(entry.Bytes), remoteAddr.String())
	err := nexus.saveBytes(remoteAddr.String(), packet.Filename)
	if err != nil {
		fmt.Fprintf(os.Stdout, "WRITE: ERROR unable to save file:[%s]", packet.Filename)
	}

	// @TODO Nullify entry
}
