package flatpack

import (
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type fixture struct {
	Foo string
	Bar []string
	Baz struct {
		Quux string
	}
}

type fixture2 struct {
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
	fixture
}

var _ = Describe("implementation", func() {
	Describe(".assign()", func() {
		It("panics over unsupported types", func() {
			it := implementation{stubEnvironment(map[string]string{})}
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
			it := implementation{stubEnvironment(map[string]string{})}
			s := badEmbedding{}
			err := it.Unmarshal(&s)
			Expect(err).To(HaveOccurred())
		})

		It("handles nested structs", func() {
			fx := fixture2{}
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

		It("allocates nested pointers only when needed", func() {
			fx := fixture2{}
			env := map[string]string{
				"FOO_BAZ": "foo baz",
			}
			it := implementation{stubEnvironment(env)}
			err := it.Unmarshal(&fx)
			Expect(err).To(Succeed())
			Expect(fx.Bar).To(BeNil())
		})
	})
})
