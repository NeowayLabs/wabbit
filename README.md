# amqputil

AMQP is a verbose protocol that makes it difficult to implement proper unit-testing on your application.
This package have an AMQP interface and two possible implementations:

* RabbitMQ
* Mock

You can use the mock implementation in your tests and use the rabbitmq implementation on your software. For example,
the code below:

```go
  mockConn := mock.Conn()       // amqputil.Conn interface
  realConn := rabbitmq.Conn()   // amqputil.Conn interface
  
  err := mockConn.Dial("amqp://localhost:5672/%2f") // will fail
  
  if err != nil {
    panic(err)
  }
  
  mock.StartRabbit()
  
  err = mockConn.Dial("amqp://localhost:5672/%2f") // now it works =D
```

It's a very straightforward implementation that need a lot of improvements yet. Take careful when using it.

[]'s
