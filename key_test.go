package flatpack_test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/xeger/flatpack"
)

func key(expr string) flatpack.Key {
	parts := strings.Split(expr, ".")
	return flatpack.Key(parts)
}

var _ = Describe("Key", func() {
	Describe(".AsEnv()", func() {
		It("separates pieces with underscore", func() {
			Expect(key("Dot.Separated").AsEnv()).To(Equal("DOT_SEPARATED"))
			Expect(key("Dot...Separated").AsEnv()).To(Equal("DOT_SEPARATED"))
		})

		It("separates CamelCase words with underscore", func() {
			Expect(key("CamelCase").AsEnv()).To(Equal("CAMEL_CASE"))
			Expect(key("CamelCASE").AsEnv()).To(Equal("CAMEL_CASE"))
		})

		It("translates non-alphanumerics to underscore", func() {
			Expect(key("weird-words-here").AsEnv()).To(Equal("WEIRD_WORDS_HERE"))
			Expect(key("weird!@#@$words#$%(*here").AsEnv()).To(Equal("WEIRD_WORDS_HERE"))
		})

	})
})
