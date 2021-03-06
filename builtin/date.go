// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package builtin

import (
	"strings"
	"time"

	. "github.com/apmckinlay/gsuneido/runtime"
	"github.com/apmckinlay/gsuneido/util/dnum"
	"github.com/apmckinlay/gsuneido/util/str"
)

type SuDateGlobal struct {
	SuBuiltin
}

func init() {
	name, ps := paramSplit(`Date(string=false, pattern=false,
		year=nil, month=nil, day=nil,
		hour=nil, minute=nil, second=nil, millisecond=nil)`)
	Global.Builtin(name, &SuDateGlobal{SuBuiltin{Fn: dateCallClass,
		BuiltinParams: BuiltinParams{ParamSpec: *ps}}})
}

func dateCallClass(_ *Thread, args []Value) Value {
	if args[0] != False && hasFields(args) {
		panic("usage: Date() or Date(string [, pattern]) or " +
			"Date(year:, month:, day:, hour:, minute:, second:)")
	}
	if args[0] != False {
		if _, ok := args[0].(SuDate); ok {
			return args[0]
		}
		var d SuDate
		s := AsStr(args[0])
		if args[1] == False {
			if strings.HasPrefix(s, "#") {
				d = DateFromLiteral(s)
			} else {
				d = ParseDate(s, "yMd")
			}
		} else {
			d = ParseDate(s, AsStr(args[1]))
		}
		if d == NilDate {
			return False
		}
		return d
	} else if hasFields(args) {
		return named(args)
	}
	return Now()
}

func hasFields(args []Value) bool {
	for i := 2; i <= 8; i++ {
		if args[i] != nil {
			return true
		}
	}
	return false
}

func named(args []Value) Value {
	now := Now()
	year := now.Year()
	month := now.Month()
	day := now.Day()
	hour := now.Hour()
	minute := now.Minute()
	second := now.Second()
	millisecond := now.Millisecond()
	if args[2] != nil {
		year = ToInt(args[2])
	}
	if args[3] != nil {
		month = ToInt(args[3])
	}
	if args[4] != nil {
		day = ToInt(args[4])
	}
	if args[5] != nil {
		hour = ToInt(args[5])
	}
	if args[6] != nil {
		minute = ToInt(args[6])
	}
	if args[7] != nil {
		second = ToInt(args[7])
	}
	if args[8] != nil {
		millisecond = ToInt(args[8])
	}
	return NormalizeDate(year, month, day, hour, minute, second, millisecond)
}

func (d *SuDateGlobal) Lookup(t *Thread, method string) Callable {
	if method == "Begin" {
		return method0(func(Value) Value { return DateFromLiteral("#17000101") })
	}
	if method == "End" {
		return method0(func(Value) Value { return DateFromLiteral("#30000101") })
	}
	return d.SuBuiltin.Lookup(t, method) // for Params
}

func (d *SuDateGlobal) String() string {
	return "Date /* builtin class */"
}

var msFactor = dnum.FromStr(".001")

func init() {
	DateMethods = Methods{
		"MinusDays": method1("(date)", func(this Value, val Value) Value {
			t1 := this.(SuDate)
			if t2, ok := val.(SuDate); ok {
				return IntVal(t1.MinusDays(t2))
			}
			panic("date.MinusDays requires date")
		}),
		"MinusSeconds": method1("(date)", func(this Value, val Value) Value {
			t1 := this.(SuDate)
			if t2, ok := val.(SuDate); ok {
				if t1.Year()-t2.Year() >= 50 {
					panic("date.MinusSeconds interval too large")
				}
				ms := t1.MinusMs(t2)
				return SuDnum{Dnum: dnum.Mul(dnum.FromInt(ms), msFactor)}
			}
			panic("date.MinusSeconds requires date")
		}),
		"FormatEn": method1("(format)", func(this, arg Value) Value {
			return SuStr(this.(SuDate).Format(ToStr(arg)))
		}),
		"GetLocalGMTBias": method0(func(this Value) Value { // should be static
			_, offset := time.Now().Zone()
			return IntVal(-offset / 60)
		}),
		"Plus": method("(years=0, months=0, days=0, "+
			"hours=0, minutes=0, seconds=0, milliseconds=0)",
			func(t *Thread, this Value, args []Value) Value {
				return this.(SuDate).Plus(ToInt(args[0]), ToInt(args[1]),
					ToInt(args[2]), ToInt(args[3]), ToInt(args[4]),
					ToInt(args[5]), ToInt(args[6]))
			}),
		"WeekDay": method1("(firstDay='Sun')", func(this, arg Value) Value {
			i := dayOfWeek(arg)
			return IntVal(((this.(SuDate).WeekDay() - i) + 7) % 7)
		}),

		"Year": method0(func(this Value) Value {
			return IntVal(this.(SuDate).Year())
		}),
		"Month": method0(func(this Value) Value {
			return IntVal(this.(SuDate).Month())
		}),
		"Day": method0(func(this Value) Value {
			return IntVal(this.(SuDate).Day())
		}),
		"Hour": method0(func(this Value) Value {
			return IntVal(this.(SuDate).Hour())
		}),
		"Minute": method0(func(this Value) Value {
			return IntVal(this.(SuDate).Minute())
		}),
		"Second": method0(func(this Value) Value {
			return IntVal(this.(SuDate).Second())
		}),
		"Millisecond": method0(func(this Value) Value {
			return IntVal(this.(SuDate).Millisecond())
		}),
	}
}

func dayOfWeek(x Value) int {
	if i, ok := x.IfInt(); ok {
		return i
	}
	s := str.ToLower(AsStr(x))
	days := []string{"sunday", "monday", "tuesday",
		"wednesday", "thursday", "friday", "saturday"}
	for i, d := range days {
		if strings.HasPrefix(d, s) {
			return i
		}
	}
	panic("usage: date.WeekDay(day name or number)")
}

var _ = builtin0("UnixTime()",
	func() Value {
		return IntVal(int(time.Now().Unix()))
	})
