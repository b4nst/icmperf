package session

import (
	"github.com/prometheus-community/pro-bing"
	"golang.org/x/sync/errgroup"
)

type Session struct {
	// pingers is a list of pingers.
	pingers []*probing.Pinger
}

// NewSession creates a new session with the given number of pingers.
func NewSession(pingers []*probing.Pinger) *Session {
	s := &Session{
		pingers: pingers,
	}
	return s
}

func (s *Session) Run() error {
	wg := new(errgroup.Group)
	for _, pinger := range s.pingers {
		p := pinger // capture range variable
		wg.Go(func() error {
			return p.Run()
		})
	}
	return wg.Wait()
}

func (s *Session) Statistics() (*Statistics, error) {
	return NewStatistics(s.pingers)
}

type StatWithSize struct {
	*probing.Statistics
	Bytes int
}
