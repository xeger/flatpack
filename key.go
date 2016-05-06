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
	if k == nil || len(k) == 0 {
		return ""
	}
	return strings.Join(k, ".")
}

// AsEnv returns this key formatted in a way that is suitable for insertion
// in the process environment.
func (k Key) AsEnv() string {
	if k == nil || len(k) == 0 {
		return ""
	}

	envKey := bytes.Buffer{}
	lastUnder := false

	for i, piece := range k {
		if i > 0 {
			// Prefix.Suffix --> PREFIX_SUFFIX
			if !lastUnder {
				envKey.WriteRune('_')
				lastUnder = true
			}
		}
		runUpper := 0
		for j, char := range piece {
			if unicode.IsUpper(char) {
				if j > 0 && !lastUnder && runUpper == 0 {
					envKey.WriteRune('_')
				}
				envKey.WriteRune(unicode.ToUpper(char))
				runUpper++
				lastUnder = false
			} else if !unicode.IsLetter(char) && !unicode.IsNumber(char) {
				if !lastUnder {
					envKey.WriteRune('_')
					lastUnder = true
					runUpper = 0
				}
			} else {
				envKey.WriteRune(unicode.ToUpper(char))
				lastUnder = false
				runUpper = 0
			}
		}
	}

	return envKey.String()
}
