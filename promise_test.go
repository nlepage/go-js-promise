package promise_test

import (
	"fmt"
	"time"

	promise "github.com/nlepage/go-js-promise"
)

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
