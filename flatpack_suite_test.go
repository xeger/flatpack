package flatpack

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

// Create a Getter that acts like a flatpack.processEnvironment but actually
// reads from a map, not from the process environment.
func stubEnvironment(pairs map[string]string) Getter {
	lookupEnv := func(key string) (string, bool) {
		value, ok := pairs[key]
		return value, ok
	}
	return processEnvironment{lookupEnv}
}

func TestFlatpack(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Flatpack Suite")
}
