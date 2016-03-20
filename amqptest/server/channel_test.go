package server

import (
	"fmt"
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
	timer := time.After(3 * time.Second)

	go func() {
		data := <-deliveries

		fmt.Printf("Got '%s' message\n", string(data.Body()))

		if string(data.Body()) != "teste" {
			t.Errorf("Failed to publish message to specified route")
			return
		}

		// Closing the channel without ack'ing
		// The message MUST be enqueued by vhost
		ch.Close()

		fmt.Printf("Closing deliveries goroutine: 122\n")
		done <- true
	}()

	select {
	case <-done:
	case <-timer:
		t.Error("No data delivered.")
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
	timer = time.After(2 * time.Second)

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
	case <-timer:
		t.Error("No data delivered.")
	}
}

// func TestAckedMessagesAreCommited(t *testing.T) {
// 	vh := NewVHost("/")

// 	ch := NewChannel(vh)

// 	err := ch.ExchangeDeclare("neoway", "topic", nil)

// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}

// 	q, err := ch.QueueDeclare("data-queue", nil)

// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}

// 	err = ch.QueueBind("data-queue", "process.data", "neoway", nil)

// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}

// 	deliveries, err := ch.Consume(
// 		q.Name(),
// 		"tag-teste",
// 		nil,
// 	)

// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}

// 	err = ch.Publish("neoway", "process.data", []byte("teste"), nil)

// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}

// 	done := make(chan bool)
// 	timer := time.After(5 * time.Second)

// 	go func() {
// 		data := <-deliveries

// 		fmt.Printf("Got '%s' message\n", string(data.Body()))

// 		if string(data.Body()) != "teste" {
// 			t.Errorf("Failed to publish message to specified route")
// 			return
// 		}

// 		data.Ack(false)
// 		done <- true
// 	}()

// 	select {
// 	case <-done:
// 	case <-timer:
// 		t.Error("No data delivered.")
// 	}

// 	// create another channel and get the same data back
// 	ch2 := NewChannel(vh)

// 	err = ch2.ExchangeDeclare("neoway", "topic", nil)

// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}

// 	q2, err := ch2.QueueDeclare("data-queue", nil)

// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}

// 	err = ch2.QueueBind("data-queue", "process.data", "neoway", nil)

// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}

// 	deliveries2, err := ch2.Consume(
// 		q2.Name(),
// 		"tag-teste",
// 		nil,
// 	)

// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}

// 	timer = time.After(2 * time.Second)

// 	go func() {
// 		data := <-deliveries2

// 		t.Error("Data ack'ed delivered again")

// 		panic("never reach here. Message was ack'ed")
// 	}()

// 	select {
// 	case <-timer:
// 	}
// }
