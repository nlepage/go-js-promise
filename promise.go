package promise

import (
	"syscall/js"
)

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
	return js.Global().Get("Promise").Call("resolve", v)
}

func Reject(v interface{}) js.Value {
	return js.Global().Get("Promise").Call("reject", v)
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

func AllSettled(ps []js.Value) []Result {
	v, _ := Await(js.Global().Get("Promise").Call("allSettled", valuesToAnys(ps)))

	results := make([]Result, 0, len(ps))
	for i := range ps {
		results = append(results, Result(v.Index(i)))
	}
	return results
}

func Any(ps []js.Value) (js.Value, error) {
	return Await(js.Global().Get("Promise").Call("any", valuesToAnys(ps)))
}

func Race(ps []js.Value) (js.Value, error) {
	return Await(js.Global().Get("Promise").Call("race", valuesToAnys(ps)))
}

type Result js.Value

func (r Result) Status() string {
	return js.Value(r).Get("status").String()
}

func (r Result) Value() js.Value {
	return js.Value(r).Get("value")
}

func (r Result) Reason() Reason {
	return Reason(js.Value(r).Get("reason"))
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

func valuesToAnys(values []js.Value) []interface{} {
	anys := make([]interface{}, len(values))
	for i, p := range values {
		anys[i] = p
	}
	return anys
}
