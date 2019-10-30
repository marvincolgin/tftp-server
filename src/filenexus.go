package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"sync"
)

// FileEntry to contain raw-bytes for file and concurency Mutex
type FileEntry struct {
	Bytes []byte
	Mutex *sync.RWMutex
}

// NewFileEntry creates the struct
func NewFileEntry() *FileEntry {
	entry := FileEntry{}
	entry.Bytes = nil
	entry.Mutex = new(sync.RWMutex)
	return &entry
}

// FileNexus is a hash-map, indexed by filename
type FileNexus struct {
	entries map[string]*FileEntry
}

// NewFileNexus create a new instance of the struct
func NewFileNexus() *FileNexus {
	return &FileNexus{
		entries: make(map[string]*FileEntry),
	}
}

// makeHashKey will create a index string based for accessing the Hashmap
func (nexus *FileNexus) makeHashKey(remoteAddr string, filename string) string {
	return fmt.Sprintf("%s$%s", remoteAddr, filename)
}

// GetEntry will retrieve the Entry for a given filename/connection (filling the file data if not loaded)
func (nexus *FileNexus) GetEntry(conn *net.UDPConn, remoteAddr, filename string) (bool, *FileEntry) {
	// since the spec denotes:
	// "Requests should be handled concurrently, but files being written to the server must not be visible until completed"
	// .. as a result, I'm taking this to mean that two clients can be using the file at the same time
	// .. this could result in Client-A reading "fileA.txt", while Client-B writes "fileA.txt"
	// .. so we are going to key our hashmap with Client+Filename

	success := false
	key := nexus.makeHashKey(remoteAddr, filename)

	// Is FILE loaded?
	if _, ok := nexus.entries[key]; ok {

		success = true

	} else {

		if fileExists(filename) {

			// Attempt to Load the FILE
			data, err := ioutil.ReadFile(filename)
			if err == nil {

				// Perform Load
				nexus.entries[key] = NewFileEntry()
				nexus.entries[key].Bytes = make([]byte, len(data))
				copy(nexus.entries[key].Bytes, data)

				success = true

			} else {

				// ERROR: unable to load
				errmsg := fmt.Sprintf("ERROR: unable to ReadFile()::Error():[%s] filename:[%s]", err.Error(), filename)
				fmt.Printf("%s\n", errmsg)
				doSendError(conn, ErrorFileNotFound, errmsg)
				conn.Close()

			}

		} else { // WRQ: New File to be Created

			nexus.entries[key] = NewFileEntry()
			nexus.entries[key].Bytes = nil // Officially, nil is correct vs 'make([]byte, 0)'
			success = true
		}
	}

	var entry *FileEntry = nil
	if success {
		entry = nexus.entries[key]
	}

	return success, entry

}

func (nexus *FileNexus) saveBytes(remoteAddr string, filename string) error {

	// Get the Key to the HashMap for entry
	key := nexus.makeHashKey(remoteAddr, filename)

	// Acquire Mutex and agree to release at end of func()
	nexus.entries[key].Mutex.Lock()
	defer nexus.entries[key].Mutex.Unlock()

	// Perform write to file
	if fileEntry, ok := nexus.entries[key]; ok {

		err := ioutil.WriteFile(filename, fileEntry.Bytes, 0644)
		if err != nil {
			return fmt.Errorf("FileNexus.saveBytes(): could not write file:[%s], err.Error():[%s]", filename, err.Error())
		}

	} else {
		return fmt.Errorf("FileNexus.saveBytes(): key could not be found in hashmap, key:[%s]", key)
	}

	return nil
}

func (nexus *FileNexus) loadBytes(filename string, rawbytes []byte) []byte {

	// Acquire (READ-ONLY) Mutext and agree to release
	nexus.entries[filename].Mutex.RLock()
	defer nexus.entries[filename].Mutex.RUnlock()

	// Load the files
	return nexus.entries[filename].Bytes
}
