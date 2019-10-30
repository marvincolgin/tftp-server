package main

import (
	"fmt"
	"os"

	"github.com/pborman/getopt"
)

func main() {

	optIP := getopt.StringLong("ip", 'i', "127.0.0.1", "Listener IP")
	optPort := getopt.IntLong("port", 'p', 69, "Listener Port")
	optThreads := getopt.IntLong("threads", 't', 16, "Max Threads")
	optTimeout := getopt.IntLong("timeout", 'o', 1, "Timeout (sec)")
	optHelp := getopt.BoolLong("help", 0, "Help")
	getopt.Parse()

	if *optHelp {
		getopt.Usage()
		os.Exit(0)
	}

	serverIPPort := fmt.Sprintf("%s:%d", *optIP, *optPort)

	ListenAndServe(serverIPPort, *optThreads, *optTimeout)
}
