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

func TestAll_rejected(t *testing.T) {
	_, err := promise.All([]js.Value{
		promise.Resolve(1),
		promise.Resolve(2),
		promise.Reject("reject all!"),
	})

	if err == nil {
		t.Fatal("all should be rejected")
	}

	if err.Error() != "reject all!" {
		t.Fatalf("all rejected with %v, expected %v", err.Error(), "reject all!")
	}
}

func ExampleAllSettled() {
	results := promise.AllSettled([]js.Value{
		promise.Resolve(1),
		promise.Reject(2),
		promise.Resolve(3),
	})

	for _, r := range results {
		switch r.Status() {
		case "fulfilled":
			fmt.Printf("resolved: %v\n", r.Value().Int())
		case "rejected":
			fmt.Printf("rejected: %v\n", r.Reason().Error())
		}
	}

	// Output:
	// resolved: 1
	// rejected: 2
	// resolved: 3
}

func ExampleRace() {
	p1, resolve1, _ := promise.New()
	p2, resolve2, _ := promise.New()

	time.AfterFunc(200*time.Millisecond, func() { resolve1("second!") })
	time.AfterFunc(100*time.Millisecond, func() { resolve2("first!") })

	v, err := promise.Race([]js.Value{p1, p2})
	if err != nil {
		fmt.Printf("error: %v\n", err.Error())
		return
	}

	fmt.Println(v.String())

	// Output:
	// first!
}

func TestRace_rejected(t *testing.T) {
	p1, resolve1, _ := promise.New()
	p2, _, reject2 := promise.New()

	time.AfterFunc(200*time.Millisecond, func() { resolve1("second!") })
	time.AfterFunc(100*time.Millisecond, func() { reject2("reject first!") })

	_, err := promise.Race([]js.Value{p1, p2})

	if err == nil {
		t.Fatal("race should be rejected")
	}

	if err.Error() != "reject first!" {
		t.Fatalf("race rejected with %v, expected %v", err.Error(), "reject first!")
	}
}

func ExampleAny() {
	p1 := promise.Reject("rejected at first!")
	p2, resolve2, _ := promise.New()
	p3, _, reject3 := promise.New()

	time.AfterFunc(200*time.Millisecond, func() { resolve2("resolved at last!") })
	time.AfterFunc(100*time.Millisecond, func() { reject3("eventually rejected!") })

	v, err := promise.Any([]js.Value{p1, p2, p3})
	if err != nil {
		fmt.Printf("error: %v\n", err.Error())
		return
	}

	fmt.Println(v.String())

	// Output:
	// resolved at last!
}

func TestAny_rejected(t *testing.T) {
	p1 := promise.Reject("rejected at first!")
	p2, _, reject2 := promise.New()
	p3, _, reject3 := promise.New()

	time.AfterFunc(200*time.Millisecond, func() { reject2("rejected at last!") })
	time.AfterFunc(100*time.Millisecond, func() { reject3("eventually rejected!") })

	_, err := promise.Any([]js.Value{p1, p2, p3})

	if err == nil {
		t.Fatal("any should be rejected")
	}

	errs := err.(promise.AggregateError).Errors()
	if len(errs) != 3 {
		t.Fatalf("any rejected with %v errors, expected %v", len(errs), 3)
	}

	if errs[0].Error() != "rejected at first!" {
		t.Fatalf("any rejected with %v, expected %v", errs[0].Error(), "rejected at first!")
	}

	if errs[1].Error() != "rejected at last!" {
		t.Fatalf("any rejected with %v, expected %v", errs[1].Error(), "rejected at last!")
	}

	if errs[2].Error() != "eventually rejected!" {
		t.Fatalf("any rejected with %v, expected %v", errs[2].Error(), "eventually rejected!")
	}
}

// FIXME TestAwait_non_promise
// FIXME TestAwait_chained
