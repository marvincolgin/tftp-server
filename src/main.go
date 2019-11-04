package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sync"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/pborman/getopt"
)

var (
	logInfo    *log.Logger
	logError   *log.Logger
	logFatal   *log.Logger
	logDebug   *log.Logger
	wg         sync.WaitGroup
	uiListener *widgets.Paragraph
	uiLog      *widgets.List
)

// Init is the package init, called automagically
func Init(logInfoH io.Writer, logErrorH io.Writer, logFatalH io.Writer, logDebugH io.Writer) {

	logInfo = log.New(logInfoH, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	logError = log.New(logErrorH, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	logFatal = log.New(logFatalH, "FATAL: ", log.Ldate|log.Ltime|log.Lshortfile)
	logDebug = log.New(logDebugH, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)

}

func uiInit(serverIPPort string) {

	// UI: Init
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}

	// UI: Listener
	uiListener = widgets.NewParagraph()
	uiListener.Text = fmt.Sprintf("Listener: %s", serverIPPort)
	uiListener.SetRect(0, 0, 25, 3)

	// UI: Log
	uiLog = widgets.NewList()
	uiLog.Title = "List"
	uiLog.Rows = []string{}
	uiLog.TextStyle = ui.NewStyle(ui.ColorYellow)
	uiLog.WrapText = false
	uiLog.SetRect(0, 7, 80, 17)

	// UI: Paint
	ui.Render(uiListener)
	ui.Render(uiLog)

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

	serverIPPort := fmt.Sprintf("%s:%d", *optIP, *optPort)

	// UI: Init
	uiInit(serverIPPort)
	defer ui.Close()

	// WORK: Start and Wait until it starts forever..loop
	wg.Add(1)
	go ListenAndServe(serverIPPort, *optThreads, *optTimeout)
	wg.Wait()

	// UI: Process Events
	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			return
		}
		ui.Render(uiLog)
	}

}
