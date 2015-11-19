package flatpack

import (
	"strings"
)

// Key is an ordered sequence of struct field names.
type Key []string

func (k Key) String() string {
	if len(k) == 0 {
		return "."
	} else {
		return strings.Join(k, ".")
	}
}
