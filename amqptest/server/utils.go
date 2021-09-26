package server

import (
	"fmt"
	"strings"
)

// matchs r2 against r1 following the AMQP rules for topic routing keys
func topicMatch(r1, r2 string) bool {
	var match bool

	bparts := strings.Split(r1, ".")
	rparts := strings.Split(r2, ".")

	if len(rparts) > len(bparts) {
		return false
	}

outer:
	for i := 0; i < len(bparts); i++ {
		bp := bparts[i]
		rp := rparts[i]

		if len(bp) == 0 {
			return false
		}

		var bsi, rsi int

		for rsi < len(rp) {
			// fmt.Printf("Testing '%c' and '%c'\n", bp[bsi], rp[rsi])

			// The char '#' matchs none or more chars (everything that is on rp[rsi])
			// next char, move on
			if bp[bsi] == '#' {
				match = true
				continue outer
			} else if bp[bsi] == '*' {
				// The '*' matchs only one character, then if it's the last char of binding part
				// and isn't the last char of rp, then surely it don't match.
				if bsi == len(bp)-1 && rsi < len(rp)-1 {
					match = false
					break outer
				}

				match = true

				if bsi < len(bp)-1 {
					bsi++
				}

				rsi++
			} else if bp[bsi] == rp[rsi] {
				// if it's the last char of binding part and it isn't an '*' or '#',
				// and it isn't the last char of rp, then we can stop here
				// because sure that route don't match the binding
				if bsi == len(bp)-1 && rsi < len(rp)-1 {
					match = false
					break outer
				}

				if bsi < len(bp)-1 {
					bsi++
				}

				rsi++

				match = true
			} else {
				match = false
				break outer
			}
		}

	}

	return match
}

// match the message headers with the bindings depending on the x-match value and ignorint headers starting with x-
func headersMatch(b BindingsMap, d *Delivery) (bool, error) {
	var cmpType string
	var init bool

	if cmpType, ok := d.Headers()["x-match"].(string); ok {
		cmpType = strings.ToLower(cmpType)
		if cmpType != "any" && cmpType != "all" {
			return false, fmt.Errorf("x-match binding should be set to \"any\" or \"all\" values. got: %s", cmpType)
		}
	} else {
		return false, fmt.Errorf("x-match binding is not set")
	}

	// If it is all the base boolean flag iteration value is true, if it is any is false
	// To simplify the return if all the iteration completes
	init = cmpType == "all"

	for key, val := range d.Headers() {
		if !strings.HasPrefix(key, "x-") {
			switch cmpType {
			case "any":
				if b.headers[key] == val.(string) {
					return true, nil
				}
			default: // case "all" to keep golint happy
				if b.headers[key] != val.(string) {
					return false, nil
				}
			}
		}
	}

	return init, nil
}
