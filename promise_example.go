package promise

import (
	"fmt"
	"time"
)

func ExampleAwait() {
	p, resolve, _ := New()

	go func() {
		time.Sleep(500 * time.Millisecond)
		resolve("waited 500ms!")
	}()

	v, _ := Await(p)

	fmt.Println(v)

	// Output:
	// waited 500ms!
}
