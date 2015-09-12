[![Build Status](https://travis-ci.org/tiago4orion/wabbit.svg?branch=master)](https://travis-ci.org/tiago4orion/wabbit) [![codecov.io](http://codecov.io/github/tiago4orion/wabbit/coverage.svg?branch=master)](http://codecov.io/github/tiago4orion/wabbit?branch=master) [![GoDoc](https://godoc.org/github.com/tiago4orion/wabbit?status.svg)](https://godoc.org/github.com/tiago4orion/wabbit)


# wabbit

> Elmer Fudd: Shhh. Be vewy vewy quiet, I'm hunting wabbits

AMQP is a verbose protocol that makes it difficult to implement proper unit-testing on your application.
The first goal of this package is provide a sane interface for an
AMQP client implementation based on the specification AMQP-0-9-1 (no extension) and then an implementation of this interface using the
well established package [streadway/amap](https://github.com/streadway/amqp) (a wrapper).

What are the advantages of this?

*Testing*

This package have an AMQP interface and two possible implementations:

* rabbitmq - Bypass to [streadway/amqp](https://github.com/streadway/amqp)
* amqptest

In the same way you can use the http package in your software and use the httptest for testing, when using wabbit is recommended to you use the rabbitmq package on your software and amqptest/client in tests. Simple example:

```go
  import (
	mockClient "github.com/tiago4orion/wabbit/amqptest/client"
	mockServer "github.com/tiago4orion/wabbit/amqptest/server"
	"github.com/tiago4orion/wabbit/rabbitmq"
  )

  ...


  mockConn := mockClient.Conn()       // amqputil.Conn interface
  realConn := rabbitmq.Conn()   // amqputil.Conn interface
  
  err := mockConn.Dial("amqp://localhost:5672/%2f") // will fail
  
  if err != nil {
    panic(err)
  }
  
  fakeServer := mock.NewServer("amqp://localhost:5672/%2f")
  fakeServer.Start()
  
  err = mockConn.Dial("amqp://localhost:5672/%2f") // now it works =D
```

It's a very straightforward implementation that need a lot of improvements yet. Take careful when using it.

KILL DA WABBIT!!!!
https://www.youtube.com/watch?v=QqC_YdG7GtM

[]'s
