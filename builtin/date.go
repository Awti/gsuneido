package builtin

import (
	"strings"

	. "github.com/apmckinlay/gsuneido/runtime"
)

type SuDateGlobal struct {
	SuBuiltin
}

func init() {
	name, ps := paramSplit(`Date(string=false, pattern=false,
		year=false, month=false, day=false,
		hour=false, minute=false, second=false, millisecond=false)`)
	Global.Add(name, &SuDateGlobal{
		SuBuiltin{dateCallClass, BuiltinParams{ParamSpec: ps}}})
}

func dateCallClass(_ *Thread, args ...Value) Value {
	if args[0] != False && hasFields(args) {
		panic("usage: Date() or Date(string [, pattern]) or " +
			"Date(year:, month:, day:, hour:, minute:, second:)")
	}
	if args[0] != False {
		if _, ok := args[0].(SuDate); ok {
			return args[0]
		}
		var d SuDate
		s := ToStr(args[0])
		if args[1] == False {
			if strings.HasPrefix(s, "#") {
				d = DateFromLiteral(s)
			} else {
				d = ParseDate(s, "yMd")
			}
		} else {
			d = ParseDate(s, ToStr(args[1]))
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
		if args[i] != False {
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
	if args[2] != False {
		year = ToInt(args[2])
	}
	if args[3] != False {
		month = ToInt(args[3])
	}
	if args[4] != False {
		day = ToInt(args[4])
	}
	if args[5] != False {
		hour = ToInt(args[5])
	}
	if args[6] != False {
		minute = ToInt(args[6])
	}
	if args[7] != False {
		second = ToInt(args[7])
	}
	if args[8] != False {
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
