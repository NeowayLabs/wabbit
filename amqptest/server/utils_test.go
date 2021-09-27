package server

import (
	"testing"

	"github.com/NeowayLabs/wabbit"
)

func matchSuccess(t *testing.T, b, r string, expected bool) {
	res := topicMatch(b, r)

	if res != expected {
		t.Errorf("route '%s' and bind '%s' returns %v", b, r, res)
	}
}

func TestTopicMatch(t *testing.T) {
	for _, pair := range []struct {
		b, r string
		ok   bool
	}{
		// OK
		{"a", "a", true},
		{"ab", "ab", true},
		{"#", "a", true},
		{"#", "aa", true},
		{"#", "aaaaaaaaaaaaaaaaaaaaaaaaa", true},
		{"#.a", "bbbb.a", true},
		{"teste#", "testeb", true},
		{"teste#", "testebbbbbbbbbbb", true},
		{"*.*", "a.b", true},
		{"*a", "aa", true},
		{"*aa", "baa", true},
		{"a*.b*", "ab.ba", true},
		{"maps.layer.stored", "maps.layer.stored", true},
		{"maps.layer.#", "maps.layer.bleh", true},

		// FAIL
		{"", "a", false},
		{"a", "b", false},
		{"*", "aa", false},
		{"a*", "aaa", false},
		{"maps.layer.*", "maps.layer.stored", false},
	} {
		matchSuccess(t, pair.b, pair.r, pair.ok)
	}
}

func TestHeadersMatch(t *testing.T) {
	q := NewQueue("test")
	for _, tt := range []struct {
		b  BindingsMap
		d  Delivery
		ok bool
	}{
		// OK
		{BindingsMap{queue: q, headers: map[string]string{}}, Delivery{headers: wabbit.Option{}}, true},

		// FAIL
		{BindingsMap{queue: q, headers: map[string]string{}}, Delivery{headers: wabbit.Option{}}, false},
	} {
		tt := tt
		t.Parallel()
		res, err := headersMatch(tt.b, &tt.d)

		if !tt.ok && err == nil {
			t.Errorf("expected error but got none")
		}

		if res != tt.ok {
			t.Errorf("expected '%v' got '%v'", tt.ok, res)
		}
	}
}
