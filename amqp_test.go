package wabbit_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/NeowayLabs/wabbit"
	"github.com/NeowayLabs/wabbit/amqptest"
	"github.com/NeowayLabs/wabbit/amqptest/server"
)

func TestBasicUsage(t *testing.T) {
	mockConn, err := amqptest.Dial("amqp://localhost:5672/%2f") // will fail

	if err == nil {
		t.Errorf("First Dial must fail because no fake server is running...")
		return
	}

	fakeServer := server.NewServer("amqp://localhost:5672/%2f")

	if fakeServer == nil {
		t.Errorf("Failed to instantiate fake server")
		return
	}

	err = fakeServer.Start()

	if err != nil {
		t.Errorf("Failed to start fake server")
	}

	mockConn, err = amqptest.Dial("amqp://localhost:5672/%2f") // now it works =D

	if err != nil {
		t.Error(err)
		return
	}

	if mockConn == nil {
		t.Error("Invalid mockConn")
		return
	}
}

func pub(conn wabbit.Conn, t *testing.T, done chan bool) {
	var (
		publisher wabbit.Publisher
		confirm   chan wabbit.Confirmation
	)

	// helper function to verify publisher confirms
	checkConfirm := func(expected uint64) error {
		c := <-confirm

		if !c.Ack() {
			return fmt.Errorf("confirmation ack should be true")
		}

		if c.DeliveryTag() != expected {
			return fmt.Errorf("confirmation delivery tag should be %d (got: %d)", expected, c.DeliveryTag())
		}

		return nil
	}

	channel, err := conn.Channel()

	if err != nil {
		t.Error(err)
		return
	}

	err = channel.ExchangeDeclare(
		"test",  // name of the exchange
		"topic", // type
		wabbit.Option{
			"durable":  true,
			"delete":   false,
			"internal": false,
			"noWait":   false,
		},
	)

	if err != nil {
		goto PubError
	}

	err = channel.Confirm(false)

	if err != nil {
		goto PubError
	}

	confirm = channel.NotifyPublish(make(chan wabbit.Confirmation, 1))

	publisher, err = amqptest.NewPublisher(conn, channel)

	if err != nil {
		goto PubError
	}

	err = publisher.Publish("test", "wabbit-test-route", []byte("msg1"), nil)

	if err != nil {
		goto PubError
	}

	err = checkConfirm(1)

	if err != nil {
		goto PubError
	}

	err = publisher.Publish("test", "wabbit-test-route", []byte("msg2"), nil)

	if err != nil {
		goto PubError
	}

	err = checkConfirm(2)

	if err != nil {
		goto PubError
	}

	err = publisher.Publish("test", "wabbit-test-route", []byte("msg3"), nil)

	if err != nil {
		goto PubError
	}

	err = checkConfirm(3)

	if err != nil {
		goto PubError
	}

	goto PubSuccess

PubError:
	t.Error(err)
PubSuccess:
	done <- true
}

func sub(conn wabbit.Conn, t *testing.T, done chan bool, bindDone chan bool) {
	var (
		err          error
		queue        wabbit.Queue
		deliveries   <-chan wabbit.Delivery
		timer        <-chan time.Time
		deliveryDone chan bool
	)

	channel, err := conn.Channel()

	if err != nil {
		goto SubError
	}

	err = channel.ExchangeDeclare(
		"test",  // name of the exchange
		"topic", // type
		wabbit.Option{
			"durable":  true,
			"delete":   false,
			"internal": false,
			"noWait":   false,
		},
	)

	if err != nil {
		goto SubError
	}

	queue, err = channel.QueueDeclare(
		"sub-queue", // name of the queue
		wabbit.Option{
			"durable":   true,
			"delete":    false,
			"exclusive": false,
			"noWait":    false,
		},
	)

	if err != nil {
		goto SubError
	}

	err = channel.QueueBind(
		queue.Name(),        // name of the queue
		"wabbit-test-route", // bindingKey
		"test",              // sourceExchange
		wabbit.Option{
			"noWait": false,
		},
	)

	if err != nil {
		goto SubError
	}

	bindDone <- true

	deliveries, err = channel.Consume(
		queue.Name(), // name
		"anyname",    // consumerTag,
		wabbit.Option{
			"noAck":     false,
			"exclusive": false,
			"noLocal":   false,
			"noWait":    false,
		},
	)

	if err != nil {
		goto SubError
	}

	timer = time.After(5 * time.Second)
	deliveryDone = make(chan bool)

	go func() {
		msg1 := <-deliveries

		if string(msg1.Body()) != "msg1" {
			t.Errorf("Unexpected message: %s", string(msg1.Body()))
			deliveryDone <- true
			return
		}

		msg2 := <-deliveries

		if string(msg2.Body()) != "msg2" {
			t.Errorf("Unexpected message: %s", string(msg2.Body()))
			deliveryDone <- true
			return
		}

		msg3 := <-deliveries

		if string(msg3.Body()) != "msg3" {
			t.Errorf("Unexpected message: %s", string(msg3.Body()))
			deliveryDone <- true
			return
		}

		deliveryDone <- true
	}()

	select {
	case <-deliveryDone:
		goto SubSuccess
	case <-timer:
		err = fmt.Errorf("No data received in sub")
		goto SubError
	}

SubError:
	t.Error(err)
SubSuccess:
	done <- true
}

func TestPubSub(t *testing.T) {
	var err error

	connString := "amqp://anyhost:anyport/%2fanyVHost"

	fakeServer := server.NewServer(connString)

	err = fakeServer.Start()

	if err != nil {
		t.Error(err)
		return
	}

	conn, err := amqptest.Dial(connString)

	if err != nil {
		t.Error(err)
		return
	}

	done := make(chan bool)
	bindingDone := make(chan bool)

	go sub(conn, t, done, bindingDone)
	<-bindingDone // AMQP will silently discards messages with no route binding.
	go pub(conn, t, done)

	<-done
	<-done
}
