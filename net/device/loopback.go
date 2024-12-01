package device

import (
	"io"
)

type loopback struct {
	queue chan []byte
}

func (l *loopback) Read(buf []byte) (int, error) {
	var err error
	payload, ok := <-l.queue
	if !ok {
		err = io.EOF
	}

	return copy(buf, payload), err
}

func (l *loopback) Write(p []byte) (int, error) {
	l.queue <- p
	return len(p), nil
}

func (l *loopback) Close() error {
	close(l.queue)
	return nil
}

// OpenLoopbackDevice opens loopback queue
func OpenLoopbackDevice(name string) (*loopback, string, error) {
	l := &loopback{
		queue: make(chan []byte),
	}

	return l, name, nil
}
