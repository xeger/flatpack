package flatpack

import (
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type simple struct {
	Foo string
	Bar []string
	Baz struct {
		Quux  string
		Corge int
	}
}

type pointery struct {
	Foo struct {
		Baz string
	}
	Bar *struct {
		Baz string
	}
}

// Test that we avoid a new panic introduced in go 1.5:
//   reflect.Value.Interface: cannot return value obtained from unexported field or method
type badEmbedding struct {
	simple
}

var _ = Describe("implementation", func() {
	Describe(".assign()", func() {
		It("panics over unsupported types", func() {
			it := implementation{stubEnvironment(map[string]string{})}
			unsupported := reflect.ValueOf(make(chan int))
			Expect(func() {
				it.assign(unsupported, "", Key{})
			}).To(Panic())
		})
	})

	Describe(".Unmarshal()", func() {
		It("propagates errors", func() {
			fx := simple{}
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
			fx := simple{}
			env := map[string]string{
				"FOO":      "foo",
				"BAZ_QUUX": "baz quux",
			}
			it := implementation{stubEnvironment(env)}
			err := it.Unmarshal(&fx)
			Expect(err).To(Succeed())
		})

		It("handles nested structs", func() {
			fx := pointery{}
			env := map[string]string{
				"FOO_BAZ": "foo baz",
				"BAR_BAZ": "bar baz",
			}
			it := implementation{stubEnvironment(env)}
			err := it.Unmarshal(&fx)
			Expect(err).To(Succeed())
			Expect(fx.Foo.Baz).To(Equal("foo baz"))
			Expect(fx.Bar).NotTo(BeNil())
			Expect(fx.Bar.Baz).To(Equal("bar baz"))
		})

		It("allocates pointers when needed", func() {
			fx := pointery{}
			env := map[string]string{
				"FOO_BAZ": "foo baz",
			}
			it := implementation{stubEnvironment(env)}
			err := it.Unmarshal(&fx)
			Expect(err).To(Succeed())
			Expect(fx.Bar).To(BeNil())

			env = map[string]string{
				"FOO_BAZ": "foo baz",
				"BAR_BAZ": "bar baz",
			}
			it = implementation{stubEnvironment(env)}
			err = it.Unmarshal(&fx)
			Expect(err).To(Succeed())
			Expect(fx.Bar).NotTo(BeNil())
		})
	})

	Context("error reporting", func() {
		It("complains about reflection without panicking", func() {
			it := implementation{stubEnvironment(map[string]string{})}
			s := badEmbedding{}
			err := it.Unmarshal(&s)
			Expect(err).To(HaveOccurred())
		})

		It("complains about malformed integers", func() {
			env := map[string]string{
				"BAZ_CORGE": "not-a-number",
			}
			it := implementation{stubEnvironment(env)}
			s := simple{}
			err := it.Unmarshal(&s)
			Expect(err).To(HaveOccurred())
		})
	})
})
