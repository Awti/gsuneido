// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package runtime

import "github.com/apmckinlay/gsuneido/util/str"

type Row []DbRec

func (row Row) Get(hdr *Header, fld string) Value {
	at, ok := hdr.Map[fld]
	if !ok || int(at.Reci) >= len(row) {
		return nil
	}
	return row[at.Reci].GetVal(int(at.Fldi))
}

func (row Row) GetRaw(hdr *Header, fld string) string {
	at, ok := hdr.Map[fld]
	if !ok || int(at.Reci) >= len(row) {
		return ""
	}
	return row[at.Reci].GetRaw(int(at.Fldi))
}

// RowAt specifies the position of a field within a Row
type RowAt struct {
	Reci int16
	Fldi int16
}

// DbRec is a Record along with its address
type DbRec struct {
	Record
	Adr int
}

// Header specifies the fields (physical) and columns (logical) for a query
type Header struct {
	Fields  [][]string
	Columns []string
	Map     map[string]RowAt
}

func (hdr *Header) EnsureMap() {
	if hdr.Map == nil { //TODO concurrency
		hdr.Map = make(map[string]RowAt, len(hdr.Fields))
		for ri, r := range hdr.Fields {
			for fi, f := range r {
				hdr.Map[f] = RowAt{int16(ri), int16(fi)}
			}
		}
	}
}

// Rules is a list of the rule columns i.e. columns that are not fields
func (hdr *Header) Rules() []string {
	rules := []string{}
	for _, col := range hdr.Columns {
		if !str.List(hdr.Fields[0]).Has(col) { //TODO handle multiple fields
			rules = append(rules, col)
		}
	}
	return rules
}
