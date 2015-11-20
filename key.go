package flatpack

import (
	"bytes"
	"strings"
	"unicode"
)

// Key is an ordered sequence of struct field names.
type Key []string

// String returns this key formatted as if were a Go expression to access
// fields of a struct, i.e. a list of dot-separated identifiers.
func (k Key) String() string {
	if len(k) == 0 {
		return "."
	}
	return strings.Join(k, ".")
}

// AsEnv returns this key formatted in a way that is suitable for insertion
// in the process environment.
func (k Key) AsEnv() string {
	envKey := bytes.Buffer{}
	lastUnder, lastUpper := false, false

	for i, piece := range k {
		if i > 0 {
			// Prefix.Suffix --> PREFIX_SUFFIX
			if !lastUnder {
				envKey.WriteRune('_')
				lastUnder = true
			}
		}
		for j, char := range piece {
			if unicode.IsUpper(char) && j > 0 {
				if !lastUnder && !lastUpper {
					envKey.WriteRune('_')
				}
				envKey.WriteRune(unicode.ToUpper(char))
				lastUpper = true
				lastUnder = false
			} else if !unicode.IsLetter(char) && !unicode.IsNumber(char) {
				if !lastUnder {
					envKey.WriteRune('_')
					lastUnder = true
				}
			} else {
				envKey.WriteRune(unicode.ToUpper(char))
				lastUpper = false
				lastUnder = false
			}
		}
	}

	return envKey.String()
}
