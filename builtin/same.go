// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package builtin

import (
	. "github.com/apmckinlay/gsuneido/runtime"
)

var _ = builtin2("Same?(x, y)",
	func(x, y Value) Value {
		return SuBool(x == y)
	})
