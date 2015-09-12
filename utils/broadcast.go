package utils

import "github.com/tiago4orion/wabbit"

const listenerSlots = 128

// ErrBroadcast enables broadcast an error channel to various listener channels
type ErrBroadcast struct {
	listeners []chan<- wabbit.Error
	c         chan wabbit.Error
}

func NewErrBroadcast() *ErrBroadcast {
	b := &ErrBroadcast{
		c:         make(chan wabbit.Error),
		listeners: make([]chan<- wabbit.Error, 0, listenerSlots),
	}

	go func() {
		for {
			select {
			case e := <-b.c:
				b.spread(e)
			}
		}
	}()

	return b
}

// Add a new listener
func (b *ErrBroadcast) Add(c chan<- wabbit.Error) {
	b.listeners = append(b.listeners, c)
}

func (b *ErrBroadcast) Write(err wabbit.Error) {
	b.c <- err
}

func (b *ErrBroadcast) spread(err wabbit.Error) {
	for _, l := range b.listeners {
		l <- err
	}
}
