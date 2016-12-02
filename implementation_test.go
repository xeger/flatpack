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
		Foo string
		Bar int
		Baz float64
	}
	Quux []int
}

type pointery struct {
	Foo struct {
		Foo string
	}
	Bar *struct {
		Foo string
	}
	Baz []*int
}

// Test that we avoid a new panic introduced in go 1.5:
//   reflect.Value.Interface: cannot return value obtained from unexported field or method
type badEmbedding struct {
	simple
}

type badType struct {
	Foo map[string]bool
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
		It("handles strings and numbers", func() {
			fx := simple{}
			env := map[string]string{
				"FOO":     "foo",
				"BAZ_FOO": "baz foo",
				"BAZ_BAR": "42",
				"BAZ_BAZ": "3.14159",
			}
			it := implementation{stubEnvironment(env)}
			err := it.Unmarshal(&fx)
			Expect(err).To(Succeed())
			Expect(fx.Foo).To(Equal("foo"))
			Expect(fx.Baz.Foo).To(Equal("baz foo"))
			Expect(fx.Baz.Bar).To(Equal(42))
			Expect(fx.Baz.Baz).To(Equal(3.14159))
		})

		It("handles slices", func() {
			fx := simple{}
			env := map[string]string{
				"BAR":  `["foo", "bar"]`,
				"QUUX": `[1,2,3]`,
			}
			it := implementation{stubEnvironment(env)}
			err := it.Unmarshal(&fx)
			Expect(err).To(Succeed())
			Expect(fx.Bar).To(Equal([]string{"foo", "bar"}))
			Expect(fx.Quux).To(Equal([]int{1, 2, 3}))
		})

		It("allocates pointers when needed", func() {
			fx := pointery{}
			env := map[string]string{
				"FOO_FOO": "foo foo",
			}
			it := implementation{stubEnvironment(env)}
			err := it.Unmarshal(&fx)
			Expect(err).To(Succeed())
			Expect(fx.Bar).To(BeNil())

			env = map[string]string{
				"FOO_FOO": "foo foo",
				"BAR_FOO": "bar foo",
				"BAZ":     `[1,2,3]`,
			}
			it = implementation{stubEnvironment(env)}
			err = it.Unmarshal(&fx)
			Expect(err).To(Succeed())
			Expect(fx.Bar).NotTo(BeNil())
			one, two, three := 1, 2, 3
			Expect(fx.Baz).To(Equal([]*int{&one, &two, &three}))
		})

		Context("error reporting", func() {
			It("complains about malformed JSON", func() {
				fx := simple{}
				env := map[string]string{
					"FOO":     "foo",
					"BAR":     `["not-a-valid-json-array`,
					"BAZ_FOO": "baz foo",
				}
				it := implementation{stubEnvironment(env)}
				err := it.Unmarshal(&fx)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(MatchRegexp("unexpected end of JSON input"))
			})

			It("complains about reflection without panicking", func() {
				it := implementation{stubEnvironment(map[string]string{})}
				s := badEmbedding{}
				err := it.Unmarshal(&s)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(MatchRegexp("reflection error"))
			})

			It("complains about malformed integers", func() {
				env := map[string]string{
					"BAZ_BAR": "not-a-number",
				}
				it := implementation{stubEnvironment(env)}
				s := simple{}
				err := it.Unmarshal(&s)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(MatchRegexp("malformed value"))
			})

			It("complains about unsupported types", func() {
				env := map[string]string{}
				it := implementation{stubEnvironment(env)}
				s := badType{}
				err := it.Unmarshal(&s)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(MatchRegexp("invalid type"))
			})
		})
	})
})
