// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package builtin

import (
	"strings"

	"github.com/apmckinlay/gsuneido/compile"
	. "github.com/apmckinlay/gsuneido/runtime"
)

var UIThread *Thread // set by main

/* builtin defines a built in function in globals
for example:
var _ = builtin("Foo(a,b)", func(t *Thread, args []Value) Value {
		...
	}))
*/
func builtin(s string, f func(t *Thread, args []Value) Value) bool {
	name, ps := paramSplit(s)
	Global.Builtin(name,
		&SuBuiltin{Fn: f, BuiltinParams: BuiltinParams{ParamSpec: *ps}})
	return true
}

func builtin0(s string, f func() Value) bool {
	name, ps := paramSplit(s)
	Global.Builtin(name,
		&SuBuiltin0{Fn: f, BuiltinParams: BuiltinParams{ParamSpec: *ps}})
	return true
}

func builtin1(s string, f func(a Value) Value) bool {
	name, ps := paramSplit(s)
	Global.Builtin(name,
		&SuBuiltin1{Fn: f, BuiltinParams: BuiltinParams{ParamSpec: *ps}})
	return true
}

func builtin2(s string, f func(a, b Value) Value) bool {
	name, ps := paramSplit(s)
	Global.Builtin(name,
		&SuBuiltin2{Fn: f, BuiltinParams: BuiltinParams{ParamSpec: *ps}})
	return true
}

func builtin3(s string, f func(a, b, c Value) Value) bool {
	name, ps := paramSplit(s)
	Global.Builtin(name,
		&SuBuiltin3{Fn: f, BuiltinParams: BuiltinParams{ParamSpec: *ps}})
	return true
}

func builtin4(s string, f func(a, b, c, d Value) Value) bool {
	name, ps := paramSplit(s)
	Global.Builtin(name,
		&SuBuiltin4{Fn: f, BuiltinParams: BuiltinParams{ParamSpec: *ps}})
	return true
}

func builtin5(s string, f func(a, b, c, d, e Value) Value) bool {
	name, ps := paramSplit(s)
	Global.Builtin(name,
		&SuBuiltin5{Fn: f, BuiltinParams: BuiltinParams{ParamSpec: *ps}})
	return true
}

func builtin6(s string, f func(a, b, c, d, e, f Value) Value) bool {
	name, ps := paramSplit(s)
	Global.Builtin(name,
		&SuBuiltin6{Fn: f, BuiltinParams: BuiltinParams{ParamSpec: *ps}})
	return true
}

func builtin7(s string, f func(a, b, c, d, e, f, g Value) Value) bool {
	name, ps := paramSplit(s)
	Global.Builtin(name,
		&SuBuiltin7{Fn: f, BuiltinParams: BuiltinParams{ParamSpec: *ps}})
	return true
}

func builtinRaw(s string, f func(t *Thread, as *ArgSpec, args []Value) Value) bool {
	name, ps := paramSplit(s)
	// params are just for documentation, SuBuiltinRaw doesn't use them
	Global.Builtin(name,
		&SuBuiltinRaw{Fn: f, BuiltinParams: BuiltinParams{ParamSpec: *ps}})
	return true
}

// paramSplit takes Foo(x, y) and returns name and ParamSpec
func paramSplit(s string) (string, *ParamSpec) {
	i := strings.IndexByte(s, byte('('))
	name := s[:i]
	ps := params(s[i:])
	ps.Name = name
	return name, ps
}

func method(p string, f func(t *Thread, this Value, args []Value) Value) Callable {
	return &SuBuiltinMethod{Fn: f,
		BuiltinParams: BuiltinParams{ParamSpec: *params(p)}}
}

func method0(f func(this Value) Value) Callable {
	return &SuBuiltinMethod0{SuBuiltin1: SuBuiltin1{Fn: f,
		BuiltinParams: BuiltinParams{}}}
}

func method1(p string, f func(this, a1 Value) Value) Callable {
	return &SuBuiltinMethod1{SuBuiltin2: SuBuiltin2{Fn: f,
		BuiltinParams: BuiltinParams{ParamSpec: *params(p)}}}
}

func method2(p string, f func(this, a1, a2 Value) Value) Callable {
	return &SuBuiltinMethod2{SuBuiltin3: SuBuiltin3{Fn: f,
		BuiltinParams: BuiltinParams{ParamSpec: *params(p)}}}
}

func method3(p string, f func(this, a1, a2, a3 Value) Value) Callable {
	return &SuBuiltinMethod3{SuBuiltin4: SuBuiltin4{Fn: f,
		BuiltinParams: BuiltinParams{ParamSpec: *params(p)}}}
}

func methodRaw(p string,
	f func(t *Thread, as *ArgSpec, this Value, args []Value) Value) Callable {
	// params are just for documentation, SuBuiltinMethodRaw doesn't use them
	return &SuBuiltinMethodRaw{Fn: f, ParamSpec: *params(p)}
}

// params builds a ParamSpec from a string like (a, b) or (@args)
func params(s string) *ParamSpec {
	s = strings.ReplaceAll(s, "nil", "'nil'")
	fn := compile.Constant("function " + s + " {}").(*SuFunc)
	for i := 0; i < int(fn.ParamSpec.Ndefaults); i++ {
		if fn.Values[i].Equal(SuStr("nil")) {
			fn.Values[i] = nil
		}
	}
	return &fn.ParamSpec
}
