package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/pborman/getopt"
)

var (
	logInfo  *log.Logger
	logError *log.Logger
	logFatal *log.Logger
	logDebug *log.Logger
)

// Init is the package init, called automagically
func Init(logInfoH io.Writer, logErrorH io.Writer, logFatalH io.Writer, logDebugH io.Writer) {

	logInfo = log.New(logInfoH, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	logError = log.New(logErrorH, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	logFatal = log.New(logFatalH, "FATAL: ", log.Ldate|log.Ltime|log.Lshortfile)
	logDebug = log.New(logDebugH, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)

}

func main() {

	// Logging: Setup
	Init(os.Stdout, os.Stderr, ioutil.Discard, ioutil.Discard)

	// Cmd-line Parameters
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

	// Server Spin-Up!
	serverIPPort := fmt.Sprintf("%s:%d", *optIP, *optPort)
	ListenAndServe(serverIPPort, *optThreads, *optTimeout)

}
