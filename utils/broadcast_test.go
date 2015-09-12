package utils

import (
	"testing"
	"time"

	"github.com/tiago4orion/wabbit"
)

func TestBroadcast(t *testing.T) {
	spread := NewErrBroadcast()

	r := make(chan wabbit.Error)
	spread.Add(r)

	go func() {
		spread.Write(NewError(1337, "teste", true, true))
	}()

	select {
	case <-time.After(5 * time.Second):
		t.Errorf("Broadcast not working....")
		return
	case v := <-r:
		if v == nil {
			t.Errorf("Invalid value for err...")
			return
		}

		if v.Code() != 1337 || v.Reason() != "teste" {
			t.Errorf("Broadcast not working...")
			return
		}
	}
}
