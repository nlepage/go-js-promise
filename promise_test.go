package promise_test

import (
	"fmt"
	"syscall/js"
	"testing"
	"time"

	promise "github.com/nlepage/go-js-promise"
)

func Example() {
	// create a new Promise
	p, resolve, reject := promise.New()

	go func() {
		time.Sleep(100 * time.Millisecond) // do some asynchronous job...

		if err := error(nil); err != nil {
			reject(err) // reject promise if something went wrong
			return
		}

		// resolve promise if all looks good
		resolve("asynchronous job is done!")
	}()

	// wait for the promise to resolve or reject
	v, err := promise.Await(p)
	if err != nil {
		fmt.Printf("error: %v\n", err.Error())
		return
	}

	fmt.Println(v)

	// Output:
	// asynchronous job is done!
}

func ExampleAll() {
	values, err := promise.All([]js.Value{
		promise.Resolve(1),
		promise.Resolve(2),
		promise.Resolve(3),
	})

	if err != nil {
		fmt.Printf("error: %v\n", err.Error())
		return
	}

	for _, v := range values {
		fmt.Println(v.Int())
	}

	// Output:
	// 1
	// 2
	// 3
}

func TestResolve(t *testing.T) {
	p := promise.Resolve("already resolved!")

	v, err := promise.Await(p)

	if !p.InstanceOf(js.Global().Get("Promise")) {
		t.Fatal("p should be instance of Promise")
	}

	if err != nil {
		t.Fatalf("p rejected with %v", err.Error())
	}

	if v.String() != "already resolved!" {
		t.Fatalf("p resolved with %v, expected %v", err.Error(), "already resolved!")
	}
}

func TestReject(t *testing.T) {
	p := promise.Reject("already rejected!")

	_, err := promise.Await(p)

	if !p.InstanceOf(js.Global().Get("Promise")) {
		t.Fatal("p should be instance of Promise")
	}

	if err == nil {
		t.Fatal("p should be rejected")
	}

	if err.Error() != "already rejected!" {
		t.Fatalf("p rejected with %v, expected %v", err.Error(), "already rejected!")
	}
}
