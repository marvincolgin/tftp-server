package main

import (
	"flag"
	"strconv"
)

var port, numThreads, timeout int
var ip string

func init() {
	flag.StringVar(&ip, "ip", "127.0.0.1", "Listener IP")
	flag.IntVar(&port, "port", 69, "Listener Port")
	flag.IntVar(&numThreads, "threads", 16, "Max Threads")
	flag.IntVar(&timeout, "timeout", 1, "Timeout (sec)")
}

func main() {

	flag.Parse()
	serverIPPort := ip + ":" + strconv.Itoa(port)

	launch(serverIPPort, numThreads, timeout)

}
