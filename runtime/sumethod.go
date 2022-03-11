// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package runtime

import "github.com/apmckinlay/gsuneido/runtime/types"

// SuMethod is a bound method originating from an SuClass or SuInstance
// when called, it sets 'this' to the origin
type SuMethod struct {
	ValueBase[SuMethod]
	fn   Value
	this Value
}

func NewSuMethod(this Value, fn Value) *SuMethod {
	return &SuMethod{fn: fn, this: this}
}

func (m *SuMethod) GetFn() Value {
	return m.fn
}

// methodForce ignores the "this" passed to Call.
// It is used by Lookup to ensure Params and Disasm get the right "this".
type methodForce struct {
	SuMethod
}

func (m *methodForce) Call(t *Thread, _ Value, as *ArgSpec) Value {
	return m.fn.Call(t, m.this, as)
}

// Value interface --------------------------------------------------

var _ Value = (*SuMethod)(nil)

func (m *SuMethod) Call(t *Thread, _ Value, as *ArgSpec) Value {
	return m.fn.Call(t, m.this, as)
}

// Lookup is used for .Params or .Disasm
func (m *SuMethod) Lookup(t *Thread, method string) Callable {
	if f := m.fn.Lookup(t, method); f != nil {
		return &methodForce{SuMethod{fn: f.(Value), this: m.fn}}
	}
	return nil
}

func (*SuMethod) Type() types.Type {
	return types.Method
}

func (m *SuMethod) String() string {
	return m.fn.String()
}

// Equal returns true if two methods have the same fn and this
func (m *SuMethod) Equal(other any) bool {
	m2, ok := other.(*SuMethod)
	return ok && *m == *m2
}

func (m *SuMethod) SetConcurrent() {
	m.this.SetConcurrent()
}

// Named interface --------------------------------------------------

var _ Named = (*SuMethod)(nil)

func (m *SuMethod) GetName() string {
	if n, ok := m.fn.(Named); ok {
		return n.GetName()
	}
	return ""
}
