// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package query

import "github.com/apmckinlay/gsuneido/util/str"

type Summarize struct {
	Query1
	by   []string
	cols []string
	ops  []string
	ons  []string
}

func (su *Summarize) String() string {
	s := su.Query1.String() + " summarize "
	if len(su.by) > 0 {
		s += str.Join(", ", su.by...) + ", "
	}
	sep := ""
	for i := range su.cols {
		s += sep
		sep = ", "
		if su.cols[i] != "" {
			s += su.cols[i] + " = "
		}
		s += su.ops[i]
		if su.ops[i] != "count" {
			s += " " + su.ons[i]
		}
	}
	return s
}

func (su *Summarize) Transform() Query {
	su.source = su.source.Transform()
	return su
}