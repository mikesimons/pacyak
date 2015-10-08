package pacsandbox

import (
	"fmt"

	"github.com/robertkrimen/otto"
)

// NotEnoughArgumentsError is used where a JS call has been marshalled to go wthout enough args
type NotEnoughArgumentsError struct {
	got      int
	expected int
	function string
}

func (e *NotEnoughArgumentsError) Error() string {
	return fmt.Sprintf("Not enough arguments provided to %s. Got %d, expected %d", e.function, e.got, e.expected)
}

func (p *PacSandbox) ottoRetValue(s interface{}, rte error) otto.Value {
	if rte != nil {
		panic(rte)
	}

	ret, err := p.vm.ToValue(s)

	if err != nil {
		panic(err)
	}

	return ret
}

func (p *PacSandbox) ottoRetString(v otto.Value, rte error) (string, error) {
	if rte != nil {
		return "", rte
	}

	ret, err := v.ToString()

	if err != nil {
		return "", err
	}

	return ret, nil
}

func (p *PacSandbox) ottoStringArgs(call otto.FunctionCall, count int, function string) []string {
	var ret []string
	for i := 0; i < count; i++ {
		if call.Argument(i).IsUndefined() {
			ret = append(ret, "")
		}

		v, err := call.Argument(i).ToString()
		if err != nil {
			panic(err)
		}
		ret = append(ret, v)
	}

	if len(ret) < count {
		panic(&NotEnoughArgumentsError{
			function: function,
			got:      len(ret),
			expected: count,
		})
	}

	return ret
}
