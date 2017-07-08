package utils

import "github.com/NeowayLabs/wabbit"

const listenerSlots = 128

// ErrBroadcast enables broadcast an error channel to various listener channels
type ErrBroadcast struct {
	listeners []chan<- wabbit.Error
	c         chan wabbit.Error
}

// NewErrBroadcast creates a broadcast object for push errors to subscribed channels
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

// Delete the listener
func (b *ErrBroadcast) Delete(c chan<- wabbit.Error) {
	i, ok := b.findIndex(c)
	if !ok {
		return
	}
	b.listeners[i] = b.listeners[len(b.listeners)-1]
	b.listeners[len(b.listeners)-1] = nil
	b.listeners = b.listeners[:len(b.listeners)-1]
}

// Write to subscribed channels
func (b *ErrBroadcast) Write(err wabbit.Error) {
	b.c <- err
}

func (b *ErrBroadcast) spread(err wabbit.Error) {
	for _, l := range b.listeners {
		l <- err
	}
}

func (b *ErrBroadcast) findIndex(c chan<- wabbit.Error) (int, bool) {
	for i := range b.listeners {
		if b.listeners[i] == c {
			return i, true
		}
	}
	return -1, false
}
