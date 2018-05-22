package main

import (
	"os/signal"
	"github.com/tehmoon/errors"
	"os"
)

var (
	ErrSigDispatcherAlreadyStarted = errors.New("SigDispatcher has already been started")
)

// Really naive signal dispatcher that registers channel for given signals.
// It will block everything is the signal is not handled as fast as possible.
type SigDispatcher struct {
	started bool
	sigHandlers map[os.Signal][]chan os.Signal
	c chan os.Signal
}

func (sd *SigDispatcher) Register(c chan os.Signal, sigs ...os.Signal) {
	for _, sig := range sigs {
		var (
			shs []chan os.Signal
			found bool
		)

		if shs, found = sd.sigHandlers[sig]; ! found {
			shs = make([]chan os.Signal, 0)
			sd.sigHandlers[sig] = shs
		}

		found = false

		for _, sh := range shs {
			if sh == c {
				found = true
				break
			}
		}

		if ! found {
			sd.sigHandlers[sig] = append(shs, c)
		}
	}
}

func (sd SigDispatcher) dispatch() {
	for {
		select {
			case sig := <- sd.c:
				if shs, found := sd.sigHandlers[sig]; found {
					for _, sh := range shs {
						sh <- sig
					}
				}
		}
	}
}

func (sd *SigDispatcher) Start() (error) {
	if sd.started {
		return ErrSigDispatcherAlreadyStarted
	}

	sigs := make([]os.Signal, 0)
	for k, _ := range sd.sigHandlers {
		sigs = append(sigs, k)
	}

	if len(sigs) > 0 {
		go sd.dispatch()
		signal.Notify(sd.c, sigs...)
	}

	sd.started = true

	return nil
}

func NewSigDispatcher() (*SigDispatcher) {
	return &SigDispatcher{
		started: false,
		sigHandlers: make(map[os.Signal][]chan os.Signal),
		c: make(chan os.Signal, 100),
	}
}
