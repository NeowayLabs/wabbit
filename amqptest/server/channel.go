package server

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"

	"github.com/NeowayLabs/wabbit"
)

type (
	Channel struct {
		*VHost

		unacked    []unackData
		muUnacked  *sync.RWMutex
		consumers  map[string]consumer
		muConsumer *sync.RWMutex

		deliveryTagCounter uint64
	}

	unackData struct {
		d wabbit.Delivery
		q *Queue
	}

	consumer struct {
		tag        string
		deliveries chan wabbit.Delivery
		done       chan bool
	}
)

var consumerSeq uint64

func uniqueConsumerTag() string {
	return fmt.Sprintf("ctag-%s-%d", os.Args[0], atomic.AddUint64(&consumerSeq, 1))
}

func NewChannel(vhost *VHost) *Channel {
	c := Channel{
		VHost:      vhost,
		unacked:    make([]unackData, 0, QueueMaxLen),
		muUnacked:  &sync.RWMutex{},
		muConsumer: &sync.RWMutex{},
		consumers:  make(map[string]consumer),
	}

	return &c
}

func (ch *Channel) Publish(exc, route string, msg []byte, _ wabbit.Option) error {
	d := NewDelivery(ch,
		msg,
		atomic.AddUint64(&ch.deliveryTagCounter, 1))

	return ch.VHost.Publish(exc, route, d, nil)
}

// Consume starts a fake consumer of queue
func (ch *Channel) Consume(queue, consumerName string, _ wabbit.Option) (<-chan wabbit.Delivery, error) {
	var (
		c consumer
	)

	if consumerName == "" {
		consumerName = uniqueConsumerTag()
	}

	c = consumer{
		tag:        consumerName,
		deliveries: make(chan wabbit.Delivery),
		done:       make(chan bool),
	}

	ch.muConsumer.RLock()

	if c2, found := ch.consumers[consumerName]; found {
		c2.done <- true
	}

	ch.consumers[consumerName] = c

	ch.muConsumer.RUnlock()

	q, ok := ch.queues[queue]

	if !ok {
		return nil, fmt.Errorf("Unknown queue '%s'", queue)
	}

	go func() {
		for {
			select {
			case <-c.done:
				close(c.deliveries)
				return
			case d := <-q.data:
				ch.addUnacked(d, q)

				// sub-select required for cases when
				// client attempts to close the channel
				// concurrently with re-enqueues of messages
				select {
				case c.deliveries <- d:
				case <-c.done:
					close(c.deliveries)
					return
				}
			}
		}
	}()

	return c.deliveries, nil
}

func (ch *Channel) addUnacked(d wabbit.Delivery, q *Queue) {
	ch.muUnacked.Lock()
	defer ch.muUnacked.Unlock()

	ch.unacked = append(ch.unacked, unackData{d, q})
}

func (ch *Channel) enqueueUnacked() {
	ch.muUnacked.Lock()
	defer ch.muUnacked.Unlock()

	for _, ud := range ch.unacked {
		ud.q.data <- ud.d
	}

	ch.unacked = make([]unackData, 0, QueueMaxLen)
}

func (ch *Channel) Ack(tag uint64, multiple bool) error {
	var (
		pos int
		ud  unackData
	)

	if !multiple {
		for pos, ud = range ch.unacked {
			if ud.d.DeliveryTag() == tag {
				break
			}
		}

		ch.muUnacked.Lock()
		ch.unacked = ch.unacked[:pos+copy(ch.unacked[pos:], ch.unacked[pos+1:])]
		ch.muUnacked.Unlock()

	} else {
		ackMessages := make([]uint64, 0, QueueMaxLen)

		for pos, ud = range ch.unacked {
			udTag := ud.d.DeliveryTag()

			if udTag <= tag {
				ackMessages = append(ackMessages, udTag)
			}
		}

		for _, udTag := range ackMessages {
			ch.Ack(udTag, false)
		}
	}

	return nil
}

func (ch *Channel) Nack(tag uint64, multiple bool, requeue bool) error {
	var (
		pos int
		ud  unackData
	)

	if !multiple {
		for pos, ud = range ch.unacked {
			if ud.d.DeliveryTag() == tag {
				break
			}
		}

		if requeue {
			ud.q.data <- ud.d
		}

		ch.muUnacked.Lock()
		ch.unacked = ch.unacked[:pos+copy(ch.unacked[pos:], ch.unacked[pos+1:])]
		ch.muUnacked.Unlock()
	} else {
		nackMessages := make([]uint64, 0, QueueMaxLen)

		for pos, ud = range ch.unacked {
			udTag := ud.d.DeliveryTag()

			if udTag <= tag {
				nackMessages = append(nackMessages, udTag)
			}
		}

		for _, udTag := range nackMessages {
			ch.Nack(udTag, false, requeue)
		}
	}

	return nil
}

func (ch *Channel) Reject(tag uint64, requeue bool) error {
	return ch.Nack(tag, false, requeue)
}

func (ch *Channel) Close() error {
	ch.muConsumer.Lock()
	defer ch.muConsumer.Unlock()

	for _, consumer := range ch.consumers {
		consumer.done <- true
	}

	ch.consumers = make(map[string]consumer)

	// enqueue shall happens only after every consumer of this channel
	// has stopped.
	ch.enqueueUnacked()

	return nil
}
