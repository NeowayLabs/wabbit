package server

import (
	"testing"
	"time"
)

func TestBasicConsumer(t *testing.T) {
	vh := NewVHost("/")

	ch := NewChannel(vh)

	err := ch.ExchangeDeclare("neoway", "topic", nil)

	if err != nil {
		t.Error(err)
		return
	}

	q, err := ch.QueueDeclare("data-queue", nil)

	if err != nil {
		t.Error(err)
		return
	}

	err = ch.QueueBind("data-queue", "process.data", "neoway", nil)

	if err != nil {
		t.Error(err)
		return
	}

	deliveries, err := ch.Consume(
		q.Name(),
		"tag-teste",
		nil,
	)

	if err != nil {
		t.Error(err)
		return
	}

	err = ch.Publish("neoway", "process.data", []byte("teste"), nil)

	if err != nil {
		t.Error(err)
		return
	}

	data := <-deliveries

	if string(data.Body()) != "teste" {
		t.Errorf("Failed to publish message to specified route")
		return
	}
}

func TestWorkerQueue(t *testing.T) {
	vh := NewVHost("/")

	ch := NewChannel(vh)

	q, err := ch.QueueDeclare("data-queue", nil)

	if err != nil {
		t.Error(err)
		return
	}

	deliveries, err := ch.Consume(
		q.Name(),
		"",
		nil,
	)

	if err != nil {
		t.Error(err)
		return
	}

	err = ch.Publish("", q.Name(), []byte("teste"), nil)

	if err != nil {
		t.Error(err)
		return
	}

	data := <-deliveries

	if string(data.Body()) != "teste" {
		t.Errorf("Failed to publish message to specified queue")
		return
	}
}

func TestUnackedMessagesArentLost(t *testing.T) {
	vh := NewVHost("/")

	ch := NewChannel(vh)

	err := ch.ExchangeDeclare("neoway", "topic", nil)

	if err != nil {
		t.Error(err)
		return
	}

	q, err := ch.QueueDeclare("data-queue", nil)

	if err != nil {
		t.Error(err)
		return
	}

	err = ch.QueueBind("data-queue", "process.data", "neoway", nil)

	if err != nil {
		t.Error(err)
		return
	}

	deliveries, err := ch.Consume(
		q.Name(),
		"tag-teste",
		nil,
	)

	if err != nil {
		t.Error(err)
		return
	}

	err = ch.Publish("neoway", "process.data", []byte("teste"), nil)

	if err != nil {
		t.Error(err)
		return
	}

	done := make(chan bool)

	go func() {
		data := <-deliveries

		if string(data.Body()) != "teste" {
			t.Errorf("Failed to publish message to specified route")
			return
		}

		// Closing the channel without ack'ing
		// The message MUST be enqueued by vhost
		ch.Close()

		done <- true
	}()

	select {
	case <-done:
	}

	// create another channel and get the same data back
	ch2 := NewChannel(vh)

	err = ch2.ExchangeDeclare("neoway", "topic", nil)

	if err != nil {
		t.Error(err)
		return
	}

	q2, err := ch2.QueueDeclare("data-queue", nil)

	if err != nil {
		t.Error(err)
		return
	}

	err = ch2.QueueBind("data-queue", "process.data", "neoway", nil)

	if err != nil {
		t.Error(err)
		return
	}

	deliveries2, err := ch2.Consume(
		q2.Name(),
		"tag-teste",
		nil,
	)

	if err != nil {
		t.Error(err)
		return
	}

	done = make(chan bool)

	go func() {
		data := <-deliveries2

		if string(data.Body()) != "teste" {
			t.Errorf("Failed to publish message to specified route")
			return
		}

		done <- true
	}()

	select {
	case <-done:
	}
}

func TestAckedMessagesAreCommited(t *testing.T) {
	vh := NewVHost("/")

	ch := NewChannel(vh)

	err := ch.ExchangeDeclare("neoway", "topic", nil)

	if err != nil {
		t.Error(err)
		return
	}

	q, err := ch.QueueDeclare("data-queue", nil)

	if err != nil {
		t.Error(err)
		return
	}

	err = ch.QueueBind("data-queue", "process.data", "neoway", nil)

	if err != nil {
		t.Error(err)
		return
	}

	deliveries, err := ch.Consume(
		q.Name(),
		"tag-teste",
		nil,
	)

	if err != nil {
		t.Error(err)
		return
	}

	err = ch.Publish("neoway", "process.data", []byte("teste"), nil)

	if err != nil {
		t.Error(err)
		return
	}

	done := make(chan bool)

	go func() {
		data := <-deliveries

		if string(data.Body()) != "teste" {
			t.Errorf("Failed to publish message to specified route")
			return
		}

		data.Ack(false)

		done <- true
	}()

	select {
	case <-done:
	}

	// Closing the old channel, this should reenqueue everything not ack'ed
	ch.Close()

	// create another channel and get the same data back
	ch2 := NewChannel(vh)

	err = ch2.ExchangeDeclare("neoway", "topic", nil)

	if err != nil {
		t.Error(err)
		return
	}

	q2, err := ch2.QueueDeclare("data-queue", nil)

	if err != nil {
		t.Error(err)
		return
	}

	err = ch2.QueueBind("data-queue", "process.data", "neoway", nil)

	if err != nil {
		t.Error(err)
		return
	}

	deliveries2, err := ch2.Consume(
		q2.Name(),
		"tag-teste",
		nil,
	)

	if err != nil {
		t.Error(err)
		return
	}

	timer := time.After(2 * time.Second)

	go func() {
		data := <-deliveries2

		t.Errorf("Data ack'ed delivered again: %s", string(data.Body()))

		panic("never reach here. Message was ack'ed")
	}()

	select {
	case <-timer:
	}
}

func TestPublishThenConsumeAck(t *testing.T) {
	vh := NewVHost("/")

	ch := NewChannel(vh)

	err := ch.ExchangeDeclare("neoway", "topic", nil)

	if err != nil {
		t.Error(err)
		return
	}

	q, err := ch.QueueDeclare("data-queue", nil)

	if err != nil {
		t.Error(err)
		return
	}

	err = ch.QueueBind("data-queue", "process.data", "neoway", nil)

	if err != nil {
		t.Error(err)
		return
	}

	err = ch.Publish("neoway", "process.data", []byte("teste"), nil)

	if err != nil {
		t.Error(err)
		return
	}

	err = ch.Close()
	if err != nil {
		t.Error(err)
		return
	}

	// start consumer
	ch2 := NewChannel(vh)

	q, err = ch2.QueueDeclare("data-queue", nil)

	if err != nil {
		t.Error(err)
		return
	}

	deliveries, err := ch2.Consume(
		q.Name(),
		"tag-teste",
		nil,
	)

	if err != nil {
		t.Error(err)
		return
	}

	done := make(chan bool)

	go func() {
		data := <-deliveries

		if string(data.Body()) != "teste" {
			t.Errorf("Failed to receive  message from published specified route")
			return
		}

		data.Ack(false)

		done <- true
	}()

	select {
	case <-done:
	}

	err = ch2.Close()

	if err != nil {
		t.Error(err)
		return
	}
}

func TestNAckedMessagesAreRequeued(t *testing.T) {
	vh := NewVHost("/")

	ch := NewChannel(vh)

	err := ch.ExchangeDeclare("neoway", "topic", nil)

	if err != nil {
		t.Error(err)
		return
	}

	q, err := ch.QueueDeclare("data-queue", nil)

	if err != nil {
		t.Error(err)
		return
	}

	err = ch.QueueBind("data-queue", "process.data", "neoway", nil)

	if err != nil {
		t.Error(err)
		return
	}

	deliveries, err := ch.Consume(
		q.Name(),
		"tag-teste",
		nil,
	)

	if err != nil {
		t.Error(err)
		return
	}

	err = ch.Publish("neoway", "process.data", []byte("teste"), nil)

	if err != nil {
		t.Error(err)
		return
	}

	done := make(chan bool)

	go func() {
		data := <-deliveries

		if string(data.Body()) != "teste" {
			t.Errorf("Failed to publish message to specified route")
			return
		}

		data.Nack(false, true)
		done <- true
	}()

	select {
	case <-done:
	}

	// Closing the old channel, this should reenqueue everything not ack'ed
	ch.Close()

	// create another channel and get the same data back
	ch2 := NewChannel(vh)

	err = ch2.ExchangeDeclare("neoway", "topic", nil)

	if err != nil {
		t.Error(err)
		return
	}

	q2, err := ch2.QueueDeclare("data-queue", nil)

	if err != nil {
		t.Error(err)
		return
	}

	err = ch2.QueueBind("data-queue", "process.data", "neoway", nil)

	if err != nil {
		t.Error(err)
		return
	}

	deliveries2, err := ch2.Consume(
		q2.Name(),
		"tag-teste",
		nil,
	)

	if err != nil {
		t.Error(err)
		return
	}

	timer := time.After(2 * time.Second)

	go func() {
		data := <-deliveries2

		if string(data.Body()) != "teste" {
			t.Error("wrong nacked message")
		}

		done <- true
	}()

	select {
	case <-done:
	case <-timer:
		t.Errorf("Data nack'ed but not redelivered")
	}
}

func TestNAckedMessagesAreRejectedWhenRequested(t *testing.T) {
	vh := NewVHost("/")

	ch := NewChannel(vh)

	err := ch.ExchangeDeclare("neoway", "topic", nil)

	if err != nil {
		t.Error(err)
		return
	}

	q, err := ch.QueueDeclare("data-queue", nil)

	if err != nil {
		t.Error(err)
		return
	}

	err = ch.QueueBind("data-queue", "process.data", "neoway", nil)

	if err != nil {
		t.Error(err)
		return
	}

	deliveries, err := ch.Consume(
		q.Name(),
		"tag-teste",
		nil,
	)

	if err != nil {
		t.Error(err)
		return
	}

	err = ch.Publish("neoway", "process.data", []byte("teste"), nil)

	if err != nil {
		t.Error(err)
		return
	}

	done := make(chan bool)

	go func() {
		data := <-deliveries

		if string(data.Body()) != "teste" {
			t.Errorf("Failed to publish message to specified route")
			return
		}

		data.Nack(false, false)
		done <- true
	}()

	select {
	case <-done:
	}

	// Closing the old channel, this should reenqueue everything not ack'ed
	ch.Close()

	// create another channel and get the same data back
	ch2 := NewChannel(vh)

	err = ch2.ExchangeDeclare("neoway", "topic", nil)

	if err != nil {
		t.Error(err)
		return
	}

	q2, err := ch2.QueueDeclare("data-queue", nil)

	if err != nil {
		t.Error(err)
		return
	}

	err = ch2.QueueBind("data-queue", "process.data", "neoway", nil)

	if err != nil {
		t.Error(err)
		return
	}

	deliveries2, err := ch2.Consume(
		q2.Name(),
		"tag-teste",
		nil,
	)

	if err != nil {
		t.Error(err)
		return
	}

	timer := time.After(2 * time.Second)

	go func() {
		<-deliveries2

		t.Error("Message shall not be redelivered")
		done <- true
	}()

	select {
	case <-done:
	case <-timer:
	}
}

// Reject

func TestRejectedMessagesAreRequeuedWhenRequested(t *testing.T) {
	vh := NewVHost("/")

	ch := NewChannel(vh)

	err := ch.ExchangeDeclare("neoway", "topic", nil)

	if err != nil {
		t.Error(err)
		return
	}

	q, err := ch.QueueDeclare("data-queue", nil)

	if err != nil {
		t.Error(err)
		return
	}

	err = ch.QueueBind("data-queue", "process.data", "neoway", nil)

	if err != nil {
		t.Error(err)
		return
	}

	deliveries, err := ch.Consume(
		q.Name(),
		"tag-teste",
		nil,
	)

	if err != nil {
		t.Error(err)
		return
	}

	err = ch.Publish("neoway", "process.data", []byte("teste"), nil)

	if err != nil {
		t.Error(err)
		return
	}

	done := make(chan bool)

	go func() {
		data := <-deliveries

		if string(data.Body()) != "teste" {
			t.Errorf("Failed to publish message to specified route")
			return
		}

		data.Reject(true)
		done <- true
	}()

	select {
	case <-done:
	}

	// Closing the old channel, this should reenqueue everything not ack'ed
	ch.Close()

	// create another channel and get the same data back
	ch2 := NewChannel(vh)

	err = ch2.ExchangeDeclare("neoway", "topic", nil)

	if err != nil {
		t.Error(err)
		return
	}

	q2, err := ch2.QueueDeclare("data-queue", nil)

	if err != nil {
		t.Error(err)
		return
	}

	err = ch2.QueueBind("data-queue", "process.data", "neoway", nil)

	if err != nil {
		t.Error(err)
		return
	}

	deliveries2, err := ch2.Consume(
		q2.Name(),
		"tag-teste",
		nil,
	)

	if err != nil {
		t.Error(err)
		return
	}

	timer := time.After(2 * time.Second)

	go func() {
		data := <-deliveries2

		if string(data.Body()) != "teste" {
			t.Error("wrong nacked message")
		}

		done <- true
	}()

	select {
	case <-done:
	case <-timer:
		t.Errorf("Data nack'ed but not redelivered")
	}
}

func TestRejectedMessagesAreRejectedWhenRequested(t *testing.T) {
	vh := NewVHost("/")

	ch := NewChannel(vh)

	err := ch.ExchangeDeclare("neoway", "topic", nil)

	if err != nil {
		t.Error(err)
		return
	}

	q, err := ch.QueueDeclare("data-queue", nil)

	if err != nil {
		t.Error(err)
		return
	}

	err = ch.QueueBind("data-queue", "process.data", "neoway", nil)

	if err != nil {
		t.Error(err)
		return
	}

	deliveries, err := ch.Consume(
		q.Name(),
		"tag-teste",
		nil,
	)

	if err != nil {
		t.Error(err)
		return
	}

	err = ch.Publish("neoway", "process.data", []byte("teste"), nil)

	if err != nil {
		t.Error(err)
		return
	}

	done := make(chan bool)

	go func() {
		data := <-deliveries

		if string(data.Body()) != "teste" {
			t.Errorf("Failed to publish message to specified route")
			return
		}

		data.Reject(false)
		done <- true
	}()

	select {
	case <-done:
	}

	// Closing the old channel, this should reenqueue everything not ack'ed
	ch.Close()

	// create another channel and get the same data back
	ch2 := NewChannel(vh)

	err = ch2.ExchangeDeclare("neoway", "topic", nil)

	if err != nil {
		t.Error(err)
		return
	}

	q2, err := ch2.QueueDeclare("data-queue", nil)

	if err != nil {
		t.Error(err)
		return
	}

	err = ch2.QueueBind("data-queue", "process.data", "neoway", nil)

	if err != nil {
		t.Error(err)
		return
	}

	deliveries2, err := ch2.Consume(
		q2.Name(),
		"tag-teste",
		nil,
	)

	if err != nil {
		t.Error(err)
		return
	}

	timer := time.After(2 * time.Second)

	go func() {
		<-deliveries2

		t.Error("Message shall not be redelivered")
		done <- true
	}()

	select {
	case <-done:
	case <-timer:
	}
}
