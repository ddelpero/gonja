package controlStructures

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/ddelpero/gonja/v2/exec"
	"github.com/ddelpero/gonja/v2/nodes"
	"github.com/ddelpero/gonja/v2/parser"
	"github.com/ddelpero/gonja/v2/tokens"
	"github.com/pkg/errors"
)

type CallControlStructure struct {
	*nodes.CallBlock
}

func (controlStructure *CallControlStructure) String() string {
	t := controlStructure.Position()
	return fmt.Sprintf("CallControlStructure(Line=%d Col=%d)", t.Line, t.Col)
}

func (controlStructure *CallControlStructure) Execute(r *exec.Renderer, tag *nodes.ControlStructureBlock) error {
	// Evaluate the macro/function to be called
	macroFunc := r.Eval(controlStructure.Macro)
	if macroFunc.IsError() {
		return errors.Wrapf(macroFunc, `Unable to evaluate macro call`)
	}

	if !macroFunc.IsCallable() {
		return errors.Errorf(`macro is not callable`)
	}

	// Create the caller function that will execute the call block body
	caller := func(params *exec.VarArgs) *exec.Value {
		var out strings.Builder
		sub := r.Inherit()
		sub.Output = &out

		// Set caller arguments in the context if any
		for i, arg := range params.Args {
			if i < len(controlStructure.CallerArgs) {
				sub.Environment.Context.Set(controlStructure.CallerArgs[i], arg)
			}
		}

		// Execute the call block body
		err := sub.ExecuteWrapper(controlStructure.Wrapper)
		if err != nil {
			return exec.AsValue(errors.Wrapf(err, `Unable to execute call block`))
		}
		return exec.AsSafeValue(out.String())
	}

	// Add the caller function to the context
	r.Environment.Context.Set("caller", caller)

	// Call the macro with VarArgs
	params := exec.NewVarArgs()
	for _, arg := range controlStructure.Args {
		value := r.Eval(arg)
		if value.IsError() {
			return errors.Wrapf(value, `Unable to evaluate call argument`)
		}
		params.Args = append(params.Args, value)
	}

	for key, arg := range controlStructure.Kwargs {
		value := r.Eval(arg)
		if value.IsError() {
			return errors.Wrapf(value, `Unable to evaluate call argument '%s'`, key)
		}
		params.KwArgs[key] = value
	}

	// Invoke the macro - macros have signature func(params *VarArgs) *Value
	results := macroFunc.Val.Call([]reflect.Value{reflect.ValueOf(params)})
	if len(results) == 0 {
		return errors.Errorf(`Macro call returned no value`)
	}

	returnValue := results[0].Interface().(*exec.Value)
	if returnValue.IsError() {
		return errors.Wrapf(returnValue, `Error calling macro`)
	}

	// Write the result to output
	_, err := fmt.Fprint(r.Output, returnValue.String())
	return err
}

func callParser(p *parser.Parser, args *parser.Parser) (nodes.ControlStructure, error) {
	callBlock := &nodes.CallBlock{
		Location:   p.Current(),
		Args:       []nodes.Expression{},
		Kwargs:     map[string]nodes.Expression{},
		CallerArgs: []string{},
	}

	// Parse the optional caller arguments: call(arg1, arg2, ...)
	if args.Match(tokens.LeftParenthesis) != nil {
		for args.Match(tokens.RightParenthesis) == nil {
			argName := args.Match(tokens.Name)
			if argName == nil {
				return nil, args.Error("Expected argument name as identifier.", nil)
			}
			callBlock.CallerArgs = append(callBlock.CallerArgs, argName.Val)

			if args.Match(tokens.RightParenthesis) != nil {
				break
			}
			if args.Match(tokens.Comma) == nil {
				return nil, args.Error("Expected ',' or ')'.", nil)
			}
		}
	}

	// Parse the macro name and arguments
	macroExpr, err := args.ParseExpression()
	if err != nil {
		return nil, err
	}
	callBlock.Macro = macroExpr

	// If it's a Call node (function call), extract the function and arguments
	if callNode, ok := macroExpr.(*nodes.Call); ok {
		callBlock.Macro = callNode.Func
		callBlock.Args = callNode.Args
		callBlock.Kwargs = callNode.Kwargs
	}

	if !args.End() {
		return nil, args.Error("Malformed call-tag.", nil)
	}

	// Wrap content until endcall
	wrapper, endargs, err := p.WrapUntil("endcall")
	if err != nil {
		return nil, err
	}
	callBlock.Wrapper = wrapper

	if !endargs.End() {
		return nil, endargs.Error("Arguments not allowed here.", nil)
	}

	return &CallControlStructure{callBlock}, nil
}
