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
		Foo  string
		Bar  int
		Baz  float64
		Quux uint
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

type ignored struct {
	Foo string
	Bar map[string]int `flatpack:"ignore"`
	baz struct {
		foo int
	} `flatpack:"ignore"`
	quux *int `flatpack:"ignore"`
}

// Test that we avoid a new panic introduced in go 1.5:
//   reflect.Value.Interface: cannot return value obtained from unexported field or method
type badEmbedding struct {
	embedded struct {
		Value int
	}
}

// Test that we avoid a new panic introduced in go 1.5:
//   reflect: reflect.Value.Set using value obtained using unexported field
type badField struct {
	value int
}

type badPointer struct {
	value *int
}

type badType struct {
	Foo map[string]bool
}

var _ = Describe("implementation", func() {
	Describe(".assign()", func() {
		It("panics over unsupported types", func() {
			it := implementation{stubEnvironment(map[string]string{})}
			unsup := reflect.ValueOf(make(chan int))
			Expect(func() {
				it.assign(unsup, "", Key{})
			}).To(Panic())

			unsup2 := reflect.ValueOf(&badType{})
			Expect(func() {
				it.assign(unsup2, "", Key{})
			}).To(Panic())
		})
	})

	Describe(".Unmarshal()", func() {
		It("handles strings and numbers", func() {
			fx := simple{}
			env := map[string]string{
				"FOO":      "foo",
				"BAZ_FOO":  "baz foo",
				"BAZ_BAR":  "42",
				"BAZ_BAZ":  "3.14159",
				"BAZ_QUUX": "42",
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

		It("ignores fields when requested to", func() {
			fx := ignored{}
			env := map[string]string{
				"FOO":  "foo",
				"BAR":  "bar",
				"BAZ":  "baz",
				"QUUX": "42",
			}
			it := implementation{stubEnvironment(env)}
			err := it.Unmarshal(&fx)
			Expect(err).To(Succeed())
			Expect(fx.Foo).To(Equal("foo"))
		})

		Context("error reporting", func() {
			It("complains about struct values", func() {
				fx := simple{}
				it := implementation{stubEnvironment(map[string]string{})}

				err := it.Unmarshal(fx)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(MatchRegexp("expected pointer to struct"))

				var fx2 *simple
				err = it.Unmarshal(fx2)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(MatchRegexp("expected non-nil pointer to struct"))
			})

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
				it := implementation{stubEnvironment(map[string]string{
					"EMBEDDED_VALUE": "42",
					"VALUE":          "43",
					"POINTER_VALUE":  "44",
				})}

				s := badEmbedding{}

				err := it.Unmarshal(&s)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(MatchRegexp("reflection error"))

				s2 := badField{}
				err = it.Unmarshal(&s2)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(MatchRegexp("reflection error"))

				s3 := badPointer{}
				err = it.Unmarshal(&s3)
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

				env = map[string]string{
					"BAZ_QUUX": "not-a-number",
				}
				it = implementation{stubEnvironment(env)}

				err = it.Unmarshal(&s)
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
