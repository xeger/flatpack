package flatpack_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/xeger/flatpack"
)

type family struct {
	Mother, Father string
	Siblings       []string
}

type person struct {
	Email   string
	Age     int
	Family  family
	Numbers []int
}

type stubGetter map[string]string

func (s stubGetter) Get(name []string) (string, error) {
	fullName := ""
	for _, v := range name {
		sep := "_"
		if fullName == "" {
			sep = ""
		}
		fullName = fmt.Sprintf("%s%s%s", fullName, sep, strings.ToUpper(v))
	}
	v, ok := s[fullName]
	if ok {
		return v, nil
	} else {
		panic(fmt.Errorf("Test failed: code tried to read unexpected key %v", fullName))
	}
}

var _ = Describe("Unmarshal", func() {
	It("works", func() {
		getter := stubGetter{
			"EMAIL":           "carol@example.com",
			"AGE":             "37",
			"FAMILY_MOTHER":   "Alice",
			"FAMILY_FATHER":   "Bob",
			"FAMILY_SIBLINGS": "[\"Dave\", \"Eve\"]",
			"NUMBERS":         "[3,7,11,42,76]",
		}

		expected := person{
			Email: "carol@example.com",
			Age:   37,
			Family: family{Mother: "Alice", Father: "Bob",
				Siblings: []string{"Dave", "Eve"}},
			Numbers: []int{3, 7, 11, 42, 76},
		}

		got := person{}
		Expect(flatpack.Unmarshal(getter, &got)).To(BeNil())

		Expect(got).To(Equal(expected))
	})
})
