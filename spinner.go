package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type Spinner struct {
	message string
	stop    chan struct{}
	done    chan struct{}
	mu      sync.Mutex
}

func NewSpinner(message string) *Spinner {
	return &Spinner{
		message: message,
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
	}
}

func (s *Spinner) Start() {
	go func() {
		defer close(s.done)
		i := 0
		for {
			select {
			case <-s.stop:
				// Clear the spinner line
				fmt.Fprintf(os.Stderr, "\r\033[K")
				return
			default:
				s.mu.Lock()
				msg := s.message
				s.mu.Unlock()
				fmt.Fprintf(os.Stderr, "\r\033[36m%s\033[0m %s", spinnerFrames[i%len(spinnerFrames)], msg)
				i++
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()
}

func (s *Spinner) Stop() {
	close(s.stop)
	<-s.done
}

func (s *Spinner) SetMessage(message string) {
	s.mu.Lock()
	s.message = message
	s.mu.Unlock()
}
