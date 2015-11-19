package flatpack

import (
	"errors"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type family struct {
	Mother, Father string
	Siblings       []string
}

type person struct {
	Email         string
	Age           uint
	Family        *family
	LuckyNumbers  []int
	Superstitious bool
	Mass          float32
}

// For testing type safety
type intPerson struct {
	Email, Age, Family, LuckyNumbers int
}
type floatPerson struct {
	Email, Age, Family, LuckyNumbers float64
}
type exoticPerson struct {
	Email map[int]bool
	Age   chan bool
}

// For testing Validater callbacks
type dysfunctionalPerson struct {
	person
}

func (dp *dysfunctionalPerson) Validate() error {
	return errors.New("Completely wrong")
}

var _ = Describe("Unmarshal()", func() {
	Context("given a processEnvironment data source", func() {
		getter := stubEnvironment(map[string]string{
			"EMAIL":           "carol@example.com",
			"AGE":             "37",
			"FAMILY_MOTHER":   "Alice",
			"FAMILY_FATHER":   "Bob",
			"FAMILY_SIBLINGS": "[\"Dave\", \"Eve\"]",
			"LUCKY_NUMBERS":   "[3,7,11,42,76]",
			"SUPERSTITIOUS":   "true",
			"MASS":            "16.84",
		})

		expected := person{
			Email: "carol@example.com",
			Age:   37,
			Family: &family{Mother: "Alice", Father: "Bob",
				Siblings: []string{"Dave", "Eve"}},
			LuckyNumbers:  []int{3, 7, 11, 42, 76},
			Superstitious: true,
			Mass:          16.84,
		}

		BeforeEach(func() { DataSource = getter })
		AfterEach(func() { DataSource = processEnvironment{os.LookupEnv} })

		It("populates the configuration", func() {
			got := person{}
			Expect(Unmarshal(&got)).To(BeNil())
			Expect(got).To(Equal(expected))
		})

		It("validates the configuration", func() {
			unhappy := dysfunctionalPerson{}
			Expect(Unmarshal(&unhappy)).To(MatchError("Completely wrong"))
		})

		It("complains about nil-pointer parameters", func() {
			var got *person
			Expect(Unmarshal(got)).To(MatchError("invalid value: need non-nil pointer"))
		})

		It("complains about non-pointer parameters", func() {
			got := person{}
			Expect(Unmarshal(got)).To(MatchError("invalid type: expected pointer-to-struct (key=.,type=struct)"))
		})

		It("complains about non-struct parameters", func() {
			wrong := "hello world"
			Expect(Unmarshal(wrong)).To(MatchError("invalid type: expected pointer-to-struct (key=.,type=string)"))
			number := 4
			wrong2 := &number
			Expect(Unmarshal(wrong2)).To(MatchError("invalid type: expected struct (key=.,type=int)"))
		})

		It("complains about type mismatches", func() {
			got := intPerson{}
			Expect(Unmarshal(&got)).To(MatchError("invalid value: cannot parse string as integer (value=carol@example.com,error=invalid syntax)"))
			got2 := floatPerson{}
			Expect(Unmarshal(&got2)).To(MatchError("invalid value: cannot parse string as float (value=carol@example.com,error=invalid syntax)"))
			got3 := exoticPerson{}
			Expect(Unmarshal(&got3)).To(MatchError("invalid type: unsupported data type (key=Email,type=map[int]bool)"))
		})
	})
})
