package main

import (
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

func (nexus *FileNexus) saveBytes(filename string, rawbytes []byte) {

	// Acquire Mutex and agree to release at end of func()
	nexus.entries[filename].Mutex.Lock()
	defer nexus.entries[filename].Mutex.Unlock()

	// Set the bytes for the file
	nexus.entries[filename].Bytes = rawbytes
}

func (nexus *FileNexus) loadBytes(filename string, rawbytes []byte) []byte {

	// Acquire (READ-ONLY) Mutext and agree to release
	nexus.entries[filename].Mutex.RLock()
	defer nexus.entries[filename].Mutex.RUnlock()

	// Load the files
	return nexus.entries[filename].Bytes
}
