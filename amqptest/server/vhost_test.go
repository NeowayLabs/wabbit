package server

import (
	"testing"

	"github.com/NeowayLabs/wabbit"
)

func TestVHostWithDefaults(t *testing.T) {
	vh := NewVHost("/")

	if vh.name != "/" {
		t.Errorf("Invalid broker name: %s", vh.name)
	}

	if len(vh.exchanges) < 5 || vh.exchanges[""] == nil ||
		vh.exchanges["amq.direct"] == nil || vh.exchanges["amq.topic"] == nil {
		t.Errorf("VHost created without the required exchanges specified by amqp 0.9.1")
	}
}

func TestQueueDeclare(t *testing.T) {
	vh := NewVHost("/")

	q, err := vh.QueueDeclare("test-queue", nil)

	if err != nil {
		t.Error(err)
	}

	if q.Name() != "test-queue" {
		t.Errorf("Invalid queue name")
	}

	if len(vh.queues) != 1 || vh.queues["test-queue"] != q {
		t.Errorf("Failed to declare queue")
	}

	if q.Messages() != 0 || q.Consumers() != 0 {
		t.Errorf("Invalid number of messages or consumers")
	}

	nwExchange, ok := vh.exchanges[""].(*DirectExchange)

	if !ok {
		t.Errorf("Exchange neoway not created")
		return
	}

	if len(nwExchange.bindings) != 1 {
		t.Errorf("Binding not created")
		return
	}
}

func TestBasicExchangeDeclare(t *testing.T) {
	vh := NewVHost("/")

	err := vh.ExchangeDeclare("neoway", "amq.direct", nil)

	if err == nil {
		t.Errorf("The exchange type correct name is direct")
		return
	}

	err = vh.ExchangeDeclare("neoway", "direct", nil)

	if err != nil {
		t.Error(err)
		return
	}

	if len(vh.exchanges) != 6 {
		t.Errorf("Exchange not properly created: %d", len(vh.exchanges))
		return
	}
}

func TestQueueBind(t *testing.T) {
	vh := NewVHost("/")

	err := vh.ExchangeDeclare("neoway", "direct", nil)

	if err != nil {
		t.Error(err)
		return
	}

	q, err := vh.QueueDeclare("queue-test", nil)

	if err != nil {
		t.Error(err)
		return
	}

	if q.Name() != "queue-test" {
		t.Errorf("Something wrong declaring queue")
		return
	}

	err = vh.QueueBind("queue-test", "process.data", "neoway", nil)

	if err != nil {
		t.Error(err)
		return
	}

	nwExchange, ok := vh.exchanges["neoway"].(*DirectExchange)

	if !ok {
		t.Errorf("Exchange neoway not created")
		return
	}

	if len(nwExchange.bindings) != 1 {
		t.Errorf("Binding not created")
		return
	}

	err = nwExchange.route("process.data", NewDelivery(&Channel{}, []byte{}, 1, wabbit.Option{}))

	if err != nil {
		t.Error(err)
		return
	}
}

func TestBasicPublish(t *testing.T) {
	vh := NewVHost("/")

	err := vh.ExchangeDeclare("neoway", "topic", nil)

	if err != nil {
		t.Error(err)
		return
	}

	q, err := vh.QueueDeclare("data", nil)

	if err != nil {
		t.Error(err)
		return
	}

	if q.Name() != "data" {
		t.Errorf("Invalid queue name")
		return
	}

	err = vh.QueueBind("data", "process.data", "neoway", nil)

	if err != nil {
		t.Error(err)
		return
	}

	err = vh.Publish("neoway", "process.data", NewDelivery(&Channel{}, []byte("teste"), 1, wabbit.Option{}), nil)

	if err != nil {
		t.Error(err)
		return
	}

	serverQueue, ok := q.(*Queue)

	if !ok {
		t.Errorf("Queue isn't of type *server.Queue")
		return
	}

	data := <-serverQueue.data

	if string(data.Body()) != "teste" {
		t.Errorf("Failed to publish message to specified route")
		return
	}
}
