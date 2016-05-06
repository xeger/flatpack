package flatpack

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"reflect"
)

type fixture struct {
	Foo string
	Bar []string
	Baz struct {
		Quux string
	}
}

// Test that we avoid a new panic introduced in go 1.5:
//   reflect.Value.Interface: cannot return value obtained from unexported field or method
type badEmbedding struct {
	fixture
}

var _ = Describe("implementation", func() {
	it := implementation{stubEnvironment(map[string]string{})}

	Describe(".assign()", func() {
		It("panics over unsupported types", func() {
			unsupported := reflect.ValueOf(make(chan int))
			Expect(func() {
				it.assign(unsupported, "")
			}).To(Panic())
		})
	})

	Describe(".Unmarshal()", func() {
		It("propagates errors", func() {
			fx := fixture{}
			env := map[string]string{
				"FOO":      "foo",
				"BAR":      "[\"not-a-valid-json-array",
				"BAZ_QUUX": "baz quux",
			}
			it := implementation{stubEnvironment(env)}
			err := it.Unmarshal(&fx)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp("unexpected end of JSON input"))
		})

		It("handles missing keys", func() {
			fx := fixture{}
			env := map[string]string{
				"FOO":      "foo",
				"BAZ_QUUX": "baz quux",
			}
			it := implementation{stubEnvironment(env)}
			err := it.Unmarshal(&fx)
			Expect(err).To(Succeed())
		})

		It("handles reflection errors without panicking", func() {
			s := badEmbedding{}
			err := it.Unmarshal(&s)
			Expect(err).To(HaveOccurred())
		})
	})
})
