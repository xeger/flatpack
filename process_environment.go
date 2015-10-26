package flatpack

import (
	"bytes"
	"unicode"
)

// A getter that reads configuration data from the process environment (or
// something similar).
type processEnvironment struct {
	lookup func(string) (string, bool)
}

func (pe processEnvironment) Get(name []string) (string, error) {
	key := pe.keyFor(name)
	value, _ := pe.lookup(key)
	return value, nil
}

// Transform a field name into an environment-variable key to be used with
// os.Getenv or similar.
func (pe processEnvironment) keyFor(name []string) string {
	key := bytes.Buffer{}
	for i, piece := range name {
		if i > 0 {
			key.WriteRune('_')
		}
		for j, char := range piece {
			if unicode.IsUpper(char) && j > 0 {
				key.WriteRune('_')
			}
			key.WriteRune(unicode.ToUpper(char))
		}
	}

	return key.String()
}
