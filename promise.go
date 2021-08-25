// Package promise is a utility for interacting with JavaScript promises.
//
// All errors returned from this package (except for promise.Any) are of type promise.Reason
// which can be converted back to js.Value if necessary.
package promise

import (
	"syscall/js"
)

const (
	// Fulfilled is the state value of a resolved promise.
	Fulfilled = "fulfilled"

	// Rejected is the state value of a rejected promise.
	Rejected = "rejected"
)

// New creates a JavaScript Promise.
//
// Returns the Promise, and the resolve/reject callback funcs.
//
// See https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Promise/Promise
func New() (p js.Value, resolve func(interface{}), reject func(interface{})) {
	var cbFunc js.Func
	cbFunc = js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		cbFunc.Release()

		resolve = func(value interface{}) {
			args[0].Invoke(value)
		}

		reject = func(value interface{}) {
			args[1].Invoke(value)
		}

		return js.Undefined()
	})

	p = js.Global().Get("Promise").New(cbFunc)

	return
}

// Resolve creates a JavaScript Promise resolved with a given value.
//
// See https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Promise/resolve
func Resolve(v interface{}) js.Value {
	return js.Global().Get("Promise").Call("resolve", v)
}

// Reject creates a JavaScript Promise rejected with a given value.
//
// See https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Promise/reject
func Reject(v interface{}) js.Value {
	return js.Global().Get("Promise").Call("reject", v)
}

// Await waits for the given Promise to be resolved or rejected.
//
// Similarly to JavaScript's await, if p isn't a Promise,
// it is returned as the result.
//
// See https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators/await
func Await(p js.Value) (js.Value, error) {
	if t := p.Type(); t != js.TypeObject && t != js.TypeFunction || p.Get("then").Type() != js.TypeFunction {
		return p, nil
	}

	resCh := make(chan js.Value)
	var then js.Func = js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		resCh <- args[0]
		return nil
	})
	defer then.Release()

	errCh := make(chan js.Value)
	var catch js.Func = js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		errCh <- args[0]
		return nil
	})
	defer catch.Release()

	p.Call("then", then).Call("catch", catch)

	select {
	case res := <-resCh:
		return res, nil
	case err := <-errCh:
		return js.Undefined(), Reason(err)
	}
}

// All waits for a given list of promises to be resolved and returns a list of the results.
//
// It returns an error if any of the input promises rejects.
//
// See https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Promise/all
func All(ps []js.Value) ([]js.Value, error) {
	v, err := Await(js.Global().Get("Promise").Call("all", valuesToAnys(ps)))
	if err != nil {
		return nil, err
	}

	values := make([]js.Value, 0, len(ps))
	for i := range ps {
		values = append(values, v.Index(i))
	}
	return values, nil
}

// AllSettled waits for a given list of promises to be fulfilled or rejected.
//
// It returns a list of promise.Result which describes the outcome of each input promises.
//
// See https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Promise/allSettled
func AllSettled(ps []js.Value) []Result {
	v, _ := Await(js.Global().Get("Promise").Call("allSettled", valuesToAnys(ps)))

	results := make([]Result, 0, len(ps))
	for i := range ps {
		results = append(results, Result(v.Index(i)))
	}
	return results
}

// Any returns the result of the first fulfilled promise in a given list of promises.
//
// If no promises in the given list fulfills, then it returns a promise.AggregateError.
//
// See https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Promise/any
func Any(ps []js.Value) (js.Value, error) {
	v, err := Await(js.Global().Get("Promise").Call("any", valuesToAnys(ps)))
	if err != nil {
		err = AggregateError{js.Value(err.(Reason))}
	}
	return v, err
}

// Race returns the result of the first fulfilled or rejected promise in a given list of promises.
//
// See https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Promise/race
func Race(ps []js.Value) (js.Value, error) {
	return Await(js.Global().Get("Promise").Call("race", valuesToAnys(ps)))
}

// Result is a JavaScript object that describes the outcome of a promise.
type Result js.Value

// Status returns the status of the promise, "fulfilled" or "rejected"
func (r Result) Status() string {
	return js.Value(r).Get("status").String()
}

// Value returns the result of the promise if status is "fulfilled", js.Undefined() otherwise.
func (r Result) Value() js.Value {
	return js.Value(r).Get("value")
}

// Reason returns the value the promise was rejected with if status is "rejected", js.Undefined() otherwise.
func (r Result) Reason() Reason {
	return Reason(js.Value(r).Get("reason"))
}

// Reason is a JavaScript value a promise was rejected with.
//
// It implements the error interface.
//
// It can be converted back to js.Value if needed.
type Reason js.Value

var _ error = Reason{}

// Error returns the string property "message" of the value if present,
// a string representation of the value otherwise.
func (r Reason) Error() string {
	v := js.Value(r)

	if v.Type() == js.TypeObject {
		if message := v.Get("message"); message.Type() == js.TypeString {
			return message.String()
		}
	}

	return js.Global().Call("String", v).String()
}

// AggregateError is a JavaScript object representing an error when several errors need to be wrapped in a single error.
//
// It is returned by promise.Any.
//
// It can be converted back to js.Error if needed.
//
// See https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/AggregateError
type AggregateError js.Error

var _ error = AggregateError{}

// Error returns the string property "message" of the error.
func (err AggregateError) Error() string {
	return js.Error(err).Error()
}

// Errors returns the list of errors wrapped by the error.
func (err AggregateError) Errors() []error {
	v := js.Error(err).Value.Get("errors")
	l := v.Length()

	errs := make([]error, l)
	for i := 0; i < l; i++ {
		errs[i] = Reason(v.Index(i))
	}
	return errs
}

// FIXME create JS array
func valuesToAnys(values []js.Value) []interface{} {
	anys := make([]interface{}, len(values))
	for i, p := range values {
		anys[i] = p
	}
	return anys
}
