# Wabbit - Go AMQP Mocking Library

[![GoDoc](https://godoc.org/github.com/NeowayLabs/wabbit?status.svg)](https://godoc.org/github.com/NeowayLabs/wabbit)
[![Go Report Card](https://goreportcard.com/badge/github.com/NeowayLabs/wabbit)](https://goreportcard.com/report/github.com/NeowayLabs/wabbit)

> Elmer Fudd: Shhh. Be vewy vewy quiet, I'm hunting wabbits

AMQP is a verbose protocol that makes it difficult to implement proper
unit-testing on your application.  The first goal of this package is
provide a sane interface for an AMQP client implementation based on
the specification AMQP-0-9-1 (no extension) and then an implementation
of this interface using the well established package [streadway/amqp](https://github.com/rabbitmq/amqp091-go) (a
wrapper).

What are the advantages of this?

### Usage

[How to use ?](/_examples)

### Testing

This package have an AMQP interface and two possible implementations:

-   wabbit/amqp - Bypass to [rabbitmq/amqp091-go](https://github.com/rabbitmq/amqp091-go)
-   wabbit/amqptest

In the same way you can use the http package in your software and use
the *httptest* for testing, when using wabbit is recommended to use the
*wabbit/amqp* package on your software and *wabbit/amqptest* in your
tests. Simple test example:

```go
package main

import (
	"testing"
	"github.com/NeowayLabs/wabbit/amqptest"
	"github.com/NeowayLabs/wabbit/amqptest/server"
	"github.com/NeowayLabs/wabbit/amqp"
)


func TestChannelCreation(t *testing.T) {
	mockConn, err := amqptest.Dial("amqp://localhost:5672/%2f") // will fail,

	if err == nil {
		t.Error("This shall fail, because no fake amqp server is running...")
	}

	fakeServer := server.NewServer("amqp://localhost:5672/%2f")
	fakeServer.Start()

	mockConn, err = amqptest.Dial("amqp://localhost:5672/%2f") // now it works =D

	if err != nil {
		t.Error(err)
	}

	//Now you can use mockConn as a real amqp connection.
	channel, err := mockConn.Channel()

    // ...
}
```
The package *amqptest/server* implements a mock AMQP server and it
can be used to simulate network partitions or broker crashs. To
create a new server instance use *server.NewServer* passing any
amqpuri. You can create more than one server, but they need to
have different amqpuris. Example below:

```go
    broker1 := server.NewServer("amqp://localhost:5672/%2f")
    broker2 := server.NewServer("amqp://192.168.10.165:5672/%2f")
    broker3 := server.NewServer("amqp://192.168.10.169:5672/%2f")

    broker1.Start()
    broker2.Start()
    broker3.Start()
```
Calling NewServer with same amqpuri will return the same server
instance.

Use *broker.Stop()* to abruptly stop the amqp server.

**There's no fake clustering support yet (maybe never)**

It's a very straightforward implementation that need a lot of
improvements yet. Take careful when using it.

[]'s
