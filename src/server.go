package main

import (
	"fmt"
	"io/ioutil"
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

	close(dataChannel)
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
			doReadReq(packetReq, nexus, conn)
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

		conn.Close()
	}
}

func doSendError(conn *net.UDPConn, code uint16, msg string) {
	p := NewPacketError(code, msg)

	conn.Write(p.Serialize())
}

func doReadReq(packet PacketRequest, nexus *FileNexus, conn *net.UDPConn) {

	// Validate OpMode
	if strings.Compare(strings.ToLower(packet.Mode), "octect") == 0 {
		errmsg := fmt.Sprintf("ERROR: mode:[%s] is not supported.\n", packet.Mode)
		fmt.Printf("%s\n", errmsg)
		doSendError(conn, ErrorNotDefined, errmsg)
		conn.Close()
		return
	}

	// Load the File into the Nexus (if not found)
	if _, ok := nexus.entries[packet.Filename]; !ok {
		data, err := ioutil.ReadFile(packet.Filename)
		if err == nil {
			fmt.Printf("err==nil\n")
			fmt.Printf("len(data):[%d]\n", len(data))

			nexus.entries[packet.Filename] = NewFileEntry()
			nexus.entries[packet.Filename].Bytes = make([]byte, len(data))
			fmt.Printf("copy")
			copy(nexus.entries[packet.Filename].Bytes, data)
			fmt.Printf("done.")
		} else {
			errmsg := fmt.Sprintf("ERROR: unable to ReadFile()::Error():[%s] filename:[%s]", err.Error(), packet.Filename)
			fmt.Printf("%s\n", errmsg)
			doSendError(conn, ErrorFileNotFound, errmsg)
			conn.Close()
			return
		}
	}

	//
	if entry, ok := nexus.entries[packet.Filename]; !ok {
		fmt.Println("len:", len(entry.Bytes))
	}

}
