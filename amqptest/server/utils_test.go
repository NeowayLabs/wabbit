package server

import "testing"

func matchSuccess(t *testing.T, b, r string, expected bool) {
	res := topicMatch(b, r)

	if res != expected {
		t.Errorf("Route '%s' and bind '%s' returns %v", b, r, res)
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
