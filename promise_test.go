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
		resolve("waited 500ms!")
	}()

	v, _ := promise.Await(p)

	fmt.Println(v)

	// Output:
	// waited 500ms!
}
