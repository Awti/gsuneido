// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package hamt

import (
	"sort"

	"github.com/apmckinlay/gsuneido/database/db19/stor"
	"github.com/apmckinlay/gsuneido/util/verify"
)

// list returns a list of the keys in the table
func (ht ItemHamt) list() []string {
	keys := make([]string, 0, 16)
	ht.ForEach(func(it *Item) {
		keys = append(keys, it.Key())
	})
	return keys
}

const blockSizeItem = 2000
const perFingerItem = 16

func (ht ItemHamt) Write(st *stor.Stor) uint64 {
	nitems := 0
	size := 2
	ht.ForEach(func(it *Item) {
		size += it.storSize()
		nitems++
	})
	if nitems == 0 {
		off, buf := st.Alloc(2)
		stor.NewWriter(buf).Put2(0)
		return off
	}
	nfingers := 1 + nitems/perFingerItem
	size += 3 * nfingers
	off, buf := st.Alloc(size)
	w := stor.NewWriter(buf)
	w.Put2(nitems)

	keys := ht.list()
	sort.Strings(keys)
	w2 := *w
	for i := 0; i < nfingers; i++ {
		w.Put3(0) // leave room
	}
	fingers := make([]int, 0, nfingers)
	for i, k := range keys {
		if i%16 == 0 {
			fingers = append(fingers, w.Len())
		}
		it,_ := ht.Get(k)
		it.Write(w)
	}
	verify.That(len(fingers) == nfingers)
	for _, f := range fingers {
		w2.Put3(f) // update with actual values
	}
	return off
}

func ReadItemHamt(st *stor.Stor, off uint64) ItemHamt {
	r := st.Reader(off)
	nitems := r.Get2()
	t := ItemHamt{}.Mutable()
	if nitems == 0 {
		return t
	}
	nfingers := 1 + nitems/perFingerItem
	for i := 0; i < nfingers; i++ {
		r.Get3() // skip the fingers
	}
	for i := 0; i < nitems; i++ {
		t.Put(ReadItem(st, r))
	}
	return t.Freeze()
}

//-------------------------------------------------------------------

type ItemPacked struct {
	stor    *stor.Stor
	off     uint64
	buf     []byte
	fingers []ItemFinger
}

type ItemFinger struct {
	table string
	pos   int
}

func NewItemPacked(st *stor.Stor, off uint64) *ItemPacked {
	buf := st.Data(off)
	r := stor.NewReader(buf)
	nitems := r.Get2()
	nfingers := 1 + nitems/perFingerItem
	fingers := make([]ItemFinger, nfingers)
	for i := 0; i < nfingers; i++ {
		fingers[i].pos = r.Get3()
	}
	for i := 0; i < nfingers; i++ {
		fingers[i].table = stor.NewReader(buf[fingers[i].pos:]).GetStr()
	}
	return &ItemPacked{stor: st, off: off, buf: buf, fingers: fingers}
}

func (p ItemPacked) Get(key string) *Item {
	pos := p.binarySearch(key)
	r := stor.NewReader(p.buf[pos:])
	count := 0
	for {
		item := ReadItem(p.stor, r)
		if item.Table == key {
			return item
		}
		count++
		if count > 20 {
			panic("linear search too long")
		}
	}
}

// binarySearch does a binary search of the fingers
func (p ItemPacked) binarySearch(table string) int {
	i, j := 0, len(p.fingers)
	count := 0
	for i < j {
		h := int(uint(i+j) >> 1) // i ≤ h < j
		if table >= p.fingers[h].table {
			i = h + 1
		} else {
			j = h
		}
		count++
		if count > 20 {
			panic("binary search too long")
		}
	}
	// i is first one greater, so we want i-1
	return int(p.fingers[i-1].pos)
}

func (p ItemPacked) Offset() uint64 {
	return p.off
}