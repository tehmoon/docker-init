package main

import (
	"syscall"
	"sync"
	"os"
)

type ProcessStatus int

const (
	PROC_STATE_NEW = iota
	PROC_STATE_REAPED
)

type ProcessCollector struct {
	sync.RWMutex
	pids map[int]*syscall.WaitStatus
}

func (pc *ProcessCollector) Start(c chan os.Signal) {
	pc.pids = make(map[int]*syscall.WaitStatus)

	go pc.collect(c)
}

func (pc *ProcessCollector) wait() {
	var ws syscall.WaitStatus
	pid, err := syscall.Wait4(-1, &ws, 0, nil)

	if err != nil {
		if err == syscall.ECHILD {
			return
		}
	}

	pc.Lock()
	defer pc.Unlock()
	pc.pids[pid] = &ws
}

func (pc *ProcessCollector) collect(c chan os.Signal) {
	for {
		select {
			case <- c:
				go pc.wait()
		}
	}
}

func (pc *ProcessCollector) GetStatus(pid int) (*syscall.WaitStatus) {
	pc.Lock()
	defer pc.Unlock()

	if ws, found := pc.pids[pid]; found {
		return ws
	}

	return nil
}
