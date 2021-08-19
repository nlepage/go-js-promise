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
		resolve("asynchronous job is done!")
	}()

	fmt.Println(p.InstanceOf(js.Global().Get("Promise")))

	v, _ := promise.Await(p)
	fmt.Println(v)

	// Output:
	// true
	// asynchronous job is done!
}

func ExampleResolve() {
	p := promise.Resolve("already resolved!")

	fmt.Println(p.InstanceOf(js.Global().Get("Promise")))

	v, _ := promise.Await(p)
	fmt.Println(v)

	// Output:
	// true
	// already resolved!
}

func ExampleReject() {
	p := promise.Reject("already rejected!")

	fmt.Println(p.InstanceOf(js.Global().Get("Promise")))

	_, err := promise.Await(p)
	fmt.Println(err)

	// Output:
	// true
	// already rejected!
}

func ExampleAwait() {
	p, resolve, _ := promise.New()

	go func() {
		time.Sleep(100 * time.Millisecond)
		resolve("resolved after 100ms!")
	}()

	v, err := promise.Await(p)
	if err != nil {
		return
	}

	fmt.Println(v)

	// Output:
	// resolved after 100ms!
}

func ExampleAwait_reject() {
	p, _, reject := promise.New()

	go func() {
		time.Sleep(100 * time.Millisecond)
		reject("rejected after 100ms!")
	}()

	_, err := promise.Await(p)
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	// rejected after 100ms!
}
