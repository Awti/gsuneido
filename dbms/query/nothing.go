// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package query

import . "github.com/apmckinlay/gsuneido/runtime"

// Nothing is a null query that produces no output.
// It results from a Join, Where, or Intersect with a Fixed conflict.
// It is not generated by the query parser.
type Nothing struct {
	queryBase
}

func NewNothing(columns []string) *Nothing {
	no := Nothing{}
	no.header = SimpleHeader(columns)
	return &no
}

var _ Query = (*Nothing)(nil)

func (*Nothing) String() string {
	return "NOTHING"
}

func (no *Nothing) Transform() Query {
	return no
}

func (*Nothing) Keys() [][]string {
	return [][]string{{}}
}

func (*Nothing) Indexes() [][]string {
	return [][]string{{}}
}

func (*Nothing) Nrows() (int, int) {
	return 0, 0
}

func (*Nothing) rowSize() int {
	return 0
}

func (*Nothing) fastSingle() bool {
	return true
}

func (*Nothing) Ordering() []string {
	return nil
}

func (*Nothing) Fixed() []Fixed {
	return nil
}

func (*Nothing) Updateable() string {
	return "nothing"
}

func (*Nothing) SingleTable() bool {
	return true
}

func (*Nothing) SetTran(QueryTran) {
}

func (*Nothing) optimize(Mode, []string, float64) (Cost, Cost, any) {
	return 0, 0, nil
}

func (*Nothing) setApproach([]string, float64, any, QueryTran) {
}

func (*Nothing) lookupCost() Cost {
	return 0
}

func (*Nothing) Lookup(*Thread, []string, []string) Row {
	return nil
}

func (*Nothing) Output(*Thread, Record) {
	panic("can't Output to nil query")
}

func (*Nothing) Get(*Thread, Dir) Row {
	return nil
}

func (*Nothing) Rewind() {
}

func (*Nothing) Select([]string, []string) {
}
