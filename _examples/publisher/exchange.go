package main

import (
	"flag"
	"github.com/NeowayLabs/wabbit"
	"github.com/NeowayLabs/wabbit/amqp"
	"log"
)

var (
	uri          = flag.String("uri", "amqp://guest:guest@localhost:5672/", "AMQP URI")
	queueName    = flag.String("queue", "test-queue", "Ephemeral AMQP queue name")
	exchange     = flag.String("exchange", "test-exchange", "Durable, non-auto-deleted AMQP exchange name")
	exchangeType = flag.String("exchange-type", "direct", "Exchange type - direct|fanout|topic|x-custom")
	body         = flag.String("body", "body test", "Body of message")
	reliable     = flag.Bool("reliable", true, "Wait for the publisher confirmation before exiting")
)

func init() {
	flag.Parse()
}

func main() {
	publish(*uri, *queueName, *exchange, *exchangeType, *body, *reliable)
}

func publish(uri string, queueName string, exchange string, exchangeType string, body string, reliable bool) {
	log.Println("[-] Connecting to", uri)
	connection, err := connect(uri)

	if err != nil {
		log.Fatalf("[x] AMQP connection error: %s", err)
	}

	log.Println("[√] Connected successfully")

	channel, err := connection.Channel()

	if err != nil {
		log.Fatalf("[x] Failed to open a channel: %s", err)
	}

	defer channel.Close()

	log.Println("[-] Declaring Exchange", exchangeType, exchange)
	err = channel.ExchangeDeclare(exchange, exchangeType, nil)

	if err != nil {
		log.Fatalf("[x] Failed to declare exchange: %s", err)
	}
	log.Println("[√] Exchange", exchange, "has been declared successfully")

	log.Println("[-] Declaring queue", queueName, "into channel")
	queue, err := declareQueue(queueName, channel)

	if err != nil {
		log.Fatalf("[x] Queue could not be declared. Error: %s", err.Error())
	}
	log.Println("[√] Queue", queueName, "has been declared successfully")

	if reliable {
		log.Printf("[-] Enabling publishing confirms.")
		if err := channel.Confirm(false); err != nil {
			log.Fatalf("[x] Channel could not be put into confirm mode: %s", err)
		}

		confirms := channel.NotifyPublish(make(chan wabbit.Confirmation, 1))

		defer confirmOne(confirms)
	}

	log.Println("[-] Sending message to queue:", queueName, "- exchange:", exchange)
	log.Println("\t", body)

	err = publishMessage(body, exchange, queue, channel)

	if err != nil {
		log.Fatalf("[x] Failed to publish a message. Error: %s", err.Error())
	}
}

func connect(uri string) (*amqp.Conn, error) {
	return amqp.Dial(uri)
}

func declareQueue(queueName string, channel wabbit.Channel) (wabbit.Queue, error) {
	return channel.QueueDeclare(
		queueName,
		wabbit.Option{
			"durable":    true,
			"autoDelete": false,
			"exclusive":  false,
			"noWait":     false,
		},
	)
}

func publishMessage(body string, exchange string, queue wabbit.Queue, channel wabbit.Channel) error {
	return channel.Publish(
		exchange,     // exchange
		queue.Name(), // routing key
		[]byte(body),
		wabbit.Option{
			"deliveryMode": 2,
			"contentType":  "text/plain",
		})
}

func confirmOne(confirms <-chan wabbit.Confirmation) {
	log.Printf("[-] Waiting for confirmation of one publishing")

	if confirmed := <-confirms; confirmed.Ack() {
		log.Printf("[√] Confirmed delivery with delivery tag: %d", confirmed.DeliveryTag)
	} else {
		log.Printf("[x] Failed delivery of delivery tag: %d", confirmed.DeliveryTag)
	}
}
