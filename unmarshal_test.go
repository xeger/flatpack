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
	Email  string
	Age    int
	Family family
}

type stubGetter map[string]string

func (s stubGetter) Get(name []string) (string, error) {
	fullName := ""
	for _, v := range(name) {
		sep := "_"
		if fullName == "" { sep = "" }
		fullName = fmt.Sprintf("%s%s%s", fullName, sep, strings.ToUpper(v))
	}
	v, _ := s[fullName]
	return v, nil
}

var _ = Describe("Unmarshal", func() {
	It("works", func() {
		getter := stubGetter{
			"EMAIL": "carol@example.com",
			"AGE": "37",
			"FAMILY_MOTHER": "Alice",
			"FAMILY_FATHER": "Bob",
			"FAMILY_SIBLINGS": "[\"Dave\", \"Eve\"]",
		}

		expected := person{
			Email: "carol@example.com",
			Age: 37,
			Family: family{Mother: "Alice", Father: "Bob", Siblings: []string{"Dave", "Eve"}},
		}

		got := person{}
		Expect(flatpack.Unmarshal(getter, got)).To(BeNil())

		Expect(got).To(Equal(expected))
	})
})
