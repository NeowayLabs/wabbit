[![Build Status](https://travis-ci.org/tiago4orion/wabbit.svg?branch=master)](https://travis-ci.org/tiago4orion/wabbit) [![codecov.io](http://codecov.io/github/tiago4orion/wabbit/coverage.svg?branch=master)](http://codecov.io/github/tiago4orion/wabbit?branch=master)


# wabbit

AMQP is a verbose protocol that makes it difficult to implement proper unit-testing on your application.
This package have an AMQP interface and two possible implementations:

* RabbitMQ - Bypass to [streadway/amqp](https://github.com/streadway/amqp)
* Mock

You can use the mock implementation in your tests and use the rabbitmq implementation on your software. For example,
the code below:

```go
  import (
	mockClient "github.com/tiago4orion/wabbit/mock/client"
	mockServer "github.com/tiago4orion/wabbit/mock/server"
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

[]'s
