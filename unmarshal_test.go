package flatpack

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type family struct {
	Mother, Father string
	Siblings       []string
}

type person struct {
	Email        string
	Age          int
	Family       family
	LuckyNumbers []int
}

// Create a Getter that acts like a flatpack.processEnvironment but actually
// reads from a map, not from the process environment.
func stubEnvironment(pairs map[string]string) Getter {
	lookupEnv := func(key string) (string, bool) {
		value, ok := pairs[key]
		return value, ok
	}
	return processEnvironment{lookupEnv}
}

var _ = Describe("Unmarshal", func() {
	Context("given a processEnvironment data source", func() {
		getter := stubEnvironment(map[string]string{
			"EMAIL":           "carol@example.com",
			"AGE":             "37",
			"FAMILY_MOTHER":   "Alice",
			"FAMILY_FATHER":   "Bob",
			"FAMILY_SIBLINGS": "[\"Dave\", \"Eve\"]",
			"LUCKY_NUMBERS":   "[3,7,11,42,76]",
		})

		BeforeEach(func() { DataSource = getter })
		AfterEach(func() { DataSource = processEnvironment{os.LookupEnv} })

		It("populates the configuration", func() {
			expected := person{
				Email: "carol@example.com",
				Age:   37,
				Family: family{Mother: "Alice", Father: "Bob",
					Siblings: []string{"Dave", "Eve"}},
				LuckyNumbers: []int{3, 7, 11, 42, 76},
			}

			got := person{}
			Expect(Unmarshal(&got)).To(BeNil())

			Expect(got).To(Equal(expected))
		})
	})
})
