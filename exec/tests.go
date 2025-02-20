package exec

import (
	"fmt"
	"slices"

	"github.com/pkg/errors"

	"github.com/ddelpero/gonja/v2/nodes"
)

// TestFunction is the type test functions must fulfil
type TestFunction func(*Context, *Value, *VarArgs) (bool, error)

func (e *Evaluator) EvalTest(expr *nodes.TestExpression) *Value {
	var value *Value
	//You can eval the expression when checking defined or undefined
	//because the point is to check if the variable is defined or not
	//and calling eval always returns an error if the variable is not defined
	if !slices.Contains(e.Environment.ExcludeEval, expr.Test.Name) {
		value = e.Eval(expr.Expression)
		if value.IsError() {
			return AsValue(errors.Wrapf(value, `Unable to evaluate expresion %s`, expr.Expression))
		}
	} else {
		//If we are checking if the variable is defined or not
		//create the value with out evaluating the expression
		value = ToValue(expr.Expression.Position().Val)
	}

	return e.ExecuteTest(expr.Test, value)
}

func (e *Evaluator) ExecuteTest(tc *nodes.TestCall, v *Value) *Value {
	params := &VarArgs{
		Args:   []*Value{},
		KwArgs: map[string]*Value{},
	}

	for _, param := range tc.Args {
		value := e.Eval(param)
		if value.IsError() {
			return AsValue(errors.Wrapf(value, `Unable to evaluate parameter %s`, param))
		}
		params.Args = append(params.Args, value)
	}

	for key, param := range tc.Kwargs {
		value := e.Eval(param)
		if value.IsError() {
			return AsValue(errors.Wrapf(value, `Unable to evaluate parameter %s`, param))
		}
		params.KwArgs[key] = value
	}

	return e.ExecuteTestByName(tc.Name, v, params)
}

func (e *Evaluator) ExecuteTestByName(name string, in *Value, params *VarArgs) *Value {
	test, ok := e.Environment.Tests.Get(name)
	if !e.Environment.Tests.Exists(name) || !ok {
		return AsValue(errors.Errorf("test '%s' not found", name))
	}

	result, err := test(e.Environment.Context, in, params)
	if callErr, ok := err.(ErrInvalidCall); ok && err != nil {
		return AsValue(fmt.Errorf("invalid call to test '%s': %s", name, callErr.Error()))
	} else if err != nil {
		return AsValue(fmt.Errorf("unable to execute test '%s': %s", name, err.Error()))
	} else {
		return AsValue(result)
	}
}
