package utils

import (
	"testing"
	"time"

	"github.com/NeowayLabs/wabbit"
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

func TestBroadcastDelete(t *testing.T) {
	tests := []struct {
		name      string
		listener  int
		placement int
	}{
		{
			name:     "delete from zero listener",
			listener: 0,
		},
		{
			name:      "delete from one listener",
			listener:  1,
			placement: 0,
		},
		{
			name:      "delete the first from 64 listeners",
			listener:  64,
			placement: 0,
		},
		{
			name:      "delete the middle from 64 listeners",
			listener:  64,
			placement: 31,
		},
		{
			name:      "delete the last from 64 listeners",
			listener:  64,
			placement: 63,
		},
		{
			name:      "delete out of bound from 64 listeners",
			listener:  64,
			placement: 64,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spread := NewErrBroadcast()
			del := make(chan wabbit.Error)

			for i := 0; i < tc.listener; i++ {
				if i == tc.placement {
					spread.Add(del)
				} else {
					spread.Add(make(chan wabbit.Error))
				}
			}
			if tc.placement < tc.listener {
				i, ok := spread.findIndex(del)
				if !ok || i != tc.placement {
					t.Errorf("add failed")
				}
			}

			spread.Delete(del)
			_, ok := spread.findIndex(del)
			if ok {
				t.Errorf("delete failed")
			}

			if tc.placement < tc.listener {
				got, want := len(spread.listeners), tc.listener-1
				if got != want {
					t.Errorf("unexpected listner size:\n- want: %d\n-  got: %d",
						want, got)
				}
			}
		})
	}
}
