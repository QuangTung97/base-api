package flock

import (
	"os"
	"sync"
	"syscall"
	"testing"
)

var mapMut sync.Mutex
var lockedFiles = map[string]*sync.Mutex{}

func getFileMutex(name string) (*sync.Mutex, bool) {
	mapMut.Lock()
	defer mapMut.Unlock()

	m, ok := lockedFiles[name]
	if !ok {
		m = &sync.Mutex{}
		lockedFiles[name] = m
	}
	return m, ok
}

// Lock should only be used for testing
func Lock(t *testing.T, filename string) {
	m, existed := getFileMutex(filename)
	m.Lock()
	t.Cleanup(m.Unlock)

	if existed {
		return
	}

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}

	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX)
	if err != nil {
		panic(err)
	}
}
