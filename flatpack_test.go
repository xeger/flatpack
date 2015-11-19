package flatpack

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"reflect"
)

var _ = Describe("flatpack", func() {
	it := flatpack{stubEnvironment(map[string]string{})}

	Describe(".assign()", func() {
		It("panics over unsupported types", func() {
			unsupported := reflect.ValueOf(make(chan int))
			Expect(func() {
				it.assign(unsupported, "")
			}).To(Panic())
		})
	})
})
