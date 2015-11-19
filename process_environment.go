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

func (pe processEnvironment) Get(name Key) (string, error) {
	key := pe.envKeyFor(name)
	value, _ := pe.lookup(key)
	return value, nil
}

// Transform a key into an environment-variable key to be used with
// os.Getenv or similar.
func (pe processEnvironment) envKeyFor(name Key) string {
	envKey := bytes.Buffer{}
	for i, piece := range name {
		if i > 0 {
			envKey.WriteRune('_')
		}
		for j, char := range piece {
			if unicode.IsUpper(char) && j > 0 {
				envKey.WriteRune('_')
			}
			envKey.WriteRune(unicode.ToUpper(char))
		}
	}

	return envKey.String()
}
