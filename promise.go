package promise

import (
	"errors"
	"syscall/js"
)

type PromiseResult struct {
	Status string
	Value  js.Value
	Reason js.Value
}

// New creates a new JavaScript Promise
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

func Resolve(v interface{}) js.Value {
	panic("not implemented")
}

func Reject(v interface{}) js.Value {
	panic("not implemented")
}

// Await waits for the Promise to be resolved and returns the value
// or an error if the promise rejected
func Await(p js.Value) (js.Value, error) {
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

func All(ps []js.Value) ([]js.Value, error) {
	return nil, errors.New("not implemented")
}

func AllSettled(ps []js.Value) ([]PromiseResult, error) {
	return nil, errors.New("not implemented")
}

func Any(ps []js.Value) (js.Value, error) {
	return js.Undefined(), errors.New("not implemented")
}

func Race(ps []js.Value) (js.Value, error) {
	return js.Undefined(), errors.New("not implemented")
}

type Reason js.Value

var _ error = Reason{}

func (r Reason) Error() string {
	v := js.Value(r)

	if v.Type() == js.TypeObject {
		if message := v.Get("message"); message.Type() == js.TypeString {
			return message.String()
		}
	}

	return js.Global().Call("String", v).String()
}
