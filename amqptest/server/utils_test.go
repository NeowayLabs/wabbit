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
		name         string
		b            BindingsMap
		d            Delivery
		expected, ok bool
	}{
		// OK
		{"no headers exchange nor message", BindingsMap{queue: q, headers: map[string]string{}}, Delivery{headers: wabbit.Option{}}, true, true},
		{"no headers exchange headers in message", BindingsMap{queue: q, headers: map[string]string{}}, Delivery{headers: wabbit.Option{"test": "test"}}, true, true},
		{"just x-match header set to all in exchange no headers in message", BindingsMap{queue: q, headers: map[string]string{"x-match": "all"}}, Delivery{headers: wabbit.Option{}}, true, true},
		{"all headers in exchange and message", BindingsMap{queue: q, headers: map[string]string{"x-match": "all", "test": "test"}}, Delivery{headers: wabbit.Option{"test": "test"}}, true, true},
		{"ignoring all \"x-\" prefixed headers", BindingsMap{queue: q, headers: map[string]string{"x-match": "all", "x-test": "test"}}, Delivery{headers: wabbit.Option{"x-test": "test"}}, true, true},
		{"any headers in exchange and message", BindingsMap{queue: q, headers: map[string]string{"x-match": "any", "test": "test", "test2": "test"}}, Delivery{headers: wabbit.Option{"test": "test"}}, true, true},

		// FAIL
		{"wrong x-match header value in bindings", BindingsMap{queue: q, headers: map[string]string{"x-match": "test"}}, Delivery{headers: wabbit.Option{}}, false, false},
		{"just x-match header set to any in exchange no headers message message", BindingsMap{queue: q, headers: map[string]string{"x-match": "any"}}, Delivery{headers: wabbit.Option{}}, false, true},
		{"all headers in exchange no headers in message", BindingsMap{queue: q, headers: map[string]string{"x-match": "all", "test": "test"}}, Delivery{headers: wabbit.Option{}}, false, true},
		{"all headers in exchange some headers in message", BindingsMap{queue: q, headers: map[string]string{"x-match": "all", "test": "test", "test2": "test"}}, Delivery{headers: wabbit.Option{"test": "test"}}, false, true},
		{"any headers in exchange and no match in message", BindingsMap{queue: q, headers: map[string]string{"x-match": "any", "test": "test"}}, Delivery{headers: wabbit.Option{"test2": "test"}}, false, true},
		{"ignoring any \"x-\" prefixed headers", BindingsMap{queue: q, headers: map[string]string{"x-match": "any", "x-test": "test"}}, Delivery{headers: wabbit.Option{"x-test": "test"}}, false, true},
		{"any headers in exchange and no match in message", BindingsMap{queue: q, headers: map[string]string{"x-match": "any", "test": "test"}}, Delivery{headers: wabbit.Option{"test2": "test"}}, false, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tt := tt
			t.Parallel()
			res, err := headersMatch(&tt.b, &tt.d)

			if !tt.ok && err == nil {
				t.Errorf("expected error but got none")
			}

			if res != tt.expected {
				t.Errorf("expected '%v' got '%v'", tt.expected, res)
			}
		})
	}
}
