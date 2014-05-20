package compile

import (
	"strings"
	"testing"

	"github.com/apmckinlay/gsuneido/interp"
	. "github.com/apmckinlay/gsuneido/util/hamcrest"
)

func TestCodegen(t *testing.T) {
	test := func(src, expected string) {
		ast := ParseFunction("function () {\n" + src + "\n}")
		fn := codegen(ast)
		da := []string{}
		var s string
		for i := 0; i < len(fn.Code); {
			i, s = interp.Disasm1(fn, i)
			da = append(da, s)
		}
		Assert(t).That(strings.Join(da, ", "), Equals(expected))
	}
	// folding
	test("1 + 2 + 3", "int 6")
	test("1 << 4", "int 16")
	test("'foo' $ 'bar'", "value 'foobar'")

	test("", "")
	test("return", "")
	test("return true", "true")
	test("true", "true")
	test("123", "int 123")
	test("a", "load a")
	test("_a", "dyload _a")
	test("G", "global G")
	test("a + b", "load a, load b, add")
	test("a is true", "load a, true, is")
	test("s = 'hello'", "value 'hello', store s")
	test("_dyn = 123", "int 123, store _dyn")
	test("a = b = c", "load c, store b, store a")
	test("n += 5", "load n, int 5, add, store n")
	test("++n", "load n, int 1, add, store n")
	test("n--", "load n, dup, int 1, sub, store n, pop")
	test("a.b", "load a, value 'b', get")
	test("a[2]", "load a, int 2, get")
	test("a.b = 123", "load a, value 'b', int 123, put")
	test("a[2] = false", "load a, int 2, false, put")
	test("a.b += 1", "load a, value 'b', dup2, get, int 1, add, put")

	Assert(t).That(func() { codegen(ParseFunction("function () { G = 1 }")) },
		Panics("invalid lvalue"))
}
