package eachers_test

import (
	"sync"
	"time"

	. "github.com/apoydence/eachers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type fakeMock struct {
	FooCalled chan bool
	FooInput  struct {
		Foo chan string
		Bar chan int
	}
}

func newFakeMock() *fakeMock {
	m := &fakeMock{}
	m.FooCalled = make(chan bool, 100)
	m.FooInput.Foo = make(chan string, 100)
	m.FooInput.Bar = make(chan int, 100)
	return m
}

func (m *fakeMock) Foo(foo string, bar int) {
	m.FooCalled <- true
	m.FooInput.Foo <- foo
	m.FooInput.Bar <- bar
}

var _ = Describe("BeCalled", func() {

	var (
		fakeMock *fakeMock
	)

	BeforeEach(func() {
		fakeMock = newFakeMock()
	})

	Describe("Expect", func() {
		Context("no method calls", func() {
			It("returns false", func() {
				Expect(fakeMock.FooInput).ToNot(BeCalled())
			})
		})

		Context("one method call", func() {
			BeforeEach(func() {
				fakeMock.Foo("foo", 2)
			})

			It("returns true for a single call", func() {
				Expect(fakeMock.FooInput).To(BeCalled())
			})

			It("returns true for a matching call", func() {
				Expect(fakeMock.FooInput).To(BeCalled(With("foo", 2)))
			})

			It("returns false for a non-matching call", func() {
				Expect(fakeMock.FooInput).ToNot(BeCalled(With("bar", 1)))
			})
		})

		Context("multiple method calls", func() {
			BeforeEach(func() {
				for i := 0; i < 5; i++ {
					fakeMock.Foo("foo", i)
				}
			})

			It("returns true for calls in the correct sequence", func() {
				Expect(fakeMock.FooInput).To(BeCalled(
					With("foo", 0),
					With("foo", 1),
					With("foo", 2),
					With("foo", 3),
					With("foo", 4),
				))
			})

			It("returns false for calls out of sequence", func() {
				Expect(fakeMock.FooInput).ToNot(BeCalled(
					With("foo", 0),
					With("foo", 1),
					With("foo", 3),
				))
			})

			It("returns false for any unmatched call", func() {
				Expect(fakeMock.FooInput).ToNot(BeCalled(
					With("foo", 0),
					With("foo", 1),
					With("bar", 2),
				))
			})
		})
	})

	Describe("Eventually/Consistently", func() {
		var (
			wg sync.WaitGroup
		)

		BeforeEach(func() {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < 5; i++ {
					fakeMock.Foo("foo", i)
					time.Sleep(10 * time.Millisecond)
				}
			}()
		})

		AfterEach(func() {
			wg.Wait()
		})

		It("returns true when a call sequence is eventually matched", func() {
			Eventually(fakeMock.FooInput).Should(BeCalled(
				With("foo", 2),
				With("foo", 3),
				With("foo", 4),
			))
		})

		It("returns false when a call sequence is not matched", func() {
			Consistently(fakeMock.FooInput).ShouldNot(BeCalled(
				With("foo", 2),
				With("foo", 4),
			))
		})
	})
})
