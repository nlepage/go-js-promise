package promise_test

import (
	"fmt"
	"syscall/js"
	"time"

	promise "github.com/nlepage/go-js-promise"
)

func ExampleNew() {
	p, resolve, reject := promise.New()

	go func() {
		// do some asynchronous job...

		if err := error(nil); err != nil {
			reject(err) // reject promise if something went wrong
			return
		}

		// resolve promise if all looks good
		resolve("asynchronous job is done")
	}()

	fmt.Println(p.InstanceOf(js.Global().Get("Promise")))

	// Output:
	// true
}

func ExampleAwait() {
	p, resolve, _ := promise.New()

	go func() {
		time.Sleep(500 * time.Millisecond)
		resolve("resolved after 500ms!")
	}()

	v, err := promise.Await(p)
	if err != nil {
		return
	}

	fmt.Println(v)

	// Output:
	// resolved after 500ms!
}

func ExampleAwait_reject() {
	p, _, reject := promise.New()

	go func() {
		time.Sleep(500 * time.Millisecond)
		reject("rejected after 500ms!")
	}()

	_, err := promise.Await(p)
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	// rejected after 500ms!
}
