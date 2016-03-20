package wabbit_test

import (
	"testing"

	"github.com/tiago4orion/wabbit/amqptest"
	"github.com/tiago4orion/wabbit/amqptest/server"
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
