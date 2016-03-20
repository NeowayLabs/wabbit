package server

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"

	"github.com/tiago4orion/wabbit"
)

type (
	Channel struct {
		*VHost
		*sync.Mutex

		unacked   []unackData
		consumers map[string]consumer
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
		VHost:     vhost,
		unacked:   make([]unackData, 0, QueueMaxLen),
		Mutex:     &sync.Mutex{},
		consumers: make(map[string]consumer),
	}

	return &c
}

// Consume starts a fake consumer of queue
func (ch *Channel) Consume(queue, consumerName string, _ wabbit.Option) (<-chan wabbit.Delivery, error) {
	var (
		found bool
		c     consumer
	)

	if consumerName == "" {
		consumerName = uniqueConsumerTag()
	}

	c = consumer{
		tag:        consumerName,
		deliveries: make(chan wabbit.Delivery),
		done:       make(chan bool),
	}

	ch.Lock()

	if c, found = ch.consumers[consumerName]; found {
		c.done <- true
	}

	ch.consumers[consumerName] = c

	ch.Unlock()

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
				c.deliveries <- d
				ch.addUnacked(d, q)
			}
		}
	}()

	return c.deliveries, nil
}

func (ch *Channel) addUnacked(d wabbit.Delivery, q *Queue) {
	ch.Lock()
	defer ch.Unlock()

	ch.unacked = append(ch.unacked, unackData{d, q})
}

func (ch *Channel) enqueueUnacked() {
	for _, ud := range ch.unacked {
		fmt.Printf("Enqueuing %s into queue %s\n", string(ud.d.Body()), ud.q.Name())
		ud.q.data <- ud.d
		fmt.Printf("Enqueued: %s\n", string(ud.d.Body()))
	}

	ch.unacked = make([]unackData, 0, QueueMaxLen)
}

func (ch *Channel) Close() error {
	ch.Lock()
	defer ch.Unlock()

	ch.enqueueUnacked()

	for name, consumer := range ch.consumers {
		fmt.Printf("Closing consumer goroutine...%s\n", name)
		consumer.done <- true
		fmt.Printf("Closed goroutine...%s\n", name)
	}

	ch.consumers = make(map[string]consumer)
	return nil
}
