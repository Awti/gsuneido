// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package btree

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/apmckinlay/gsuneido/db19/index/ixkey"
	"github.com/apmckinlay/gsuneido/db19/stor"
	"github.com/apmckinlay/gsuneido/runtime"
	"github.com/apmckinlay/gsuneido/util/cksum"
)

type T = btree

// btree is an immutable btree designed to be stored in a file.
type btree struct {
	// treeLevels is how many levels of tree nodes there are (initially 0)
	// Nodes do not store whether they are leaf or tree nodes.
	// Since we always start at the root and descend,
	// the code tracks the depth and compares it to treeLevels
	// to differentiate leaf or tree nodes.
	// When the root splits, treeLevels is incremented.
	treeLevels int
	// root is the offset of the root node
	root uint64
	// stor is where the btree is stored
	stor *stor.Stor
	// ixspec is an opaque value passed to GetLeafKey.
	// It specifies which fields make up the key, based on the schema.
	ixspec *ixkey.Spec
}

const maxlevels = 8

// MaxNodeSize is the maximum node size in bytes, split if larger.
// Overridden by tests.
var MaxNodeSize = 256 //TODO tune

// GetLeafKey is used to get the key for a data offset.
// It is a dependency that must be injected
var GetLeafKey func(st *stor.Stor, is *ixkey.Spec, off uint64) string

func CreateBtree(st *stor.Stor, is *ixkey.Spec) *btree {
	rootNode := node{}
	root := rootNode.putNode(st)
	return &btree{root: root, stor: st, ixspec: is}
}

func OpenBtree(st *stor.Stor, root uint64, treeLevels int) *btree {
	return &btree{root: root, treeLevels: treeLevels, stor: st}
}

func (bt *btree) GetIxspec() *ixkey.Spec {
	return bt.ixspec
}

func (bt *btree) SetIxspec(is *ixkey.Spec) {
	bt.ixspec = is
}

func (bt *btree) getLeafKey(off uint64) string {
	return GetLeafKey(bt.stor, bt.ixspec, off)
}

// Lookup returns the offset for a key, or 0 if not found.
func (bt *btree) Lookup(key string) uint64 {
	off := bt.root
	for i := 0; i <= bt.treeLevels; i++ {
		nd := bt.getNode(off)
		off = nd.search(key)
	}
	if bt.getLeafKey(off) != key {
		return 0
	}
	return off
}

// putNode stores the node
func (nd node) putNode(st *stor.Stor) uint64 {
	n := len(nd)
	off, buf := st.Alloc(2 + n + cksum.Len)
	stor.NewWriter(buf).Put2(n)
	buf = buf[2:]
	copy(buf, nd)
	cksum.Update(buf)
	// if len(nd) > 0 && rand.Intn(500) == 42 {
	// 	// corrupt some nodes to test checking
	// 	fmt.Println("ZAP")
	// 	buf := st.Data(off)
	// 	buf[3 + rand.Intn(len(nd))] = byte(rand.Intn(256))
	// }
	return off
}

// getNode returns the node for a given offset
func (bt *btree) getNode(off uint64) node {
	return readNode(bt.stor, off)
}

func (bt *btree) getNodeCk(off uint64, check bool) node {
	nd := readNode(bt.stor, off)
	if check {
		cksum.MustCheck(nd[:len(nd)+cksum.Len])
	}
	return nd
}

func readNode(st *stor.Stor, off uint64) node {
	buf := st.Data(off)
	n := stor.NewReader(buf).Get2()
	return node(buf[2 : 2+n])
}

//-------------------------------------------------------------------
// Quick check is used when opening a database. It should be fast.
// To be fast it should only look at the end (recent) part of the file.

// recentSize is the length of the tail of the file that we look at
const recentSize = 32 * 1024 * 1024 // ???

func (bt *btree) QuickCheck() {
	recent := int64(bt.stor.Size()) - recentSize
	bt.quickCheck1(0, bt.root, recent)
}

func (bt *btree) quickCheck1(depth int, offset uint64, recent int64) {
	// only look at nodes in the recent part of the file
	if int64(offset) < recent {
		return
	}
	nd := bt.getNodeCk(offset, true)
	if depth < bt.treeLevels {
		// tree node
		for it := nd.iter(); it.next(); {
			bt.quickCheck1(depth+1, it.offset, recent)
		}
	} else {
		// leaf node
		for it := nd.iter(); it.next(); {
			// only checksum data records in the recent part of the file
			if int64(it.offset) > recent {
				buf := bt.stor.Data(it.offset)
				size := runtime.RecLen(buf)
				cksum.MustCheck(buf[:size+cksum.Len])
			}
		}
	}
}

// Check verifies that the keys are in order and returns the number of keys.
// If the supplied function is not nil, it is applied to each leaf offset.
func (bt *btree) Check(fn func(uint64)) (count, size, nnodes int) {
	key := ""
	return bt.check1(0, bt.root, &key, fn)
}

func (bt *btree) check1(depth int, offset uint64, key *string,
	fn func(uint64)) (count, size, nnodes int) {
	nd := bt.getNodeCk(offset, true)
	size += len(nd)
	nnodes++
	for it := nd.iter(); it.next(); {
		offset := it.offset
		if depth < bt.treeLevels {
			// tree
			if it.pos > 0 && *key > string(it.known) {
				panic("keys out of order")
			}
			*key = string(it.known)
			c, s, n := bt.check1(depth+1, offset, key, fn) // RECURSE
			count += c
			size += s
			nnodes += n
		} else {
			// leaf
			count++
			if fn != nil {
				fn(offset)
			}
			itkey := bt.getLeafKey(offset)
			if !strings.HasPrefix(itkey, string(it.known)) {
				panic("index key does not match data")
			}
			if *key > itkey {
				panic("keys out of order")
			}
			*key = itkey
		}
	}
	return
}

// print ------------------------------------------------------------

func (bt *btree) print() {
	fmt.Println("<<<------------------------------")
	bt.print1(0, bt.root)
	fmt.Println("------------------------------>>>")
}

func (bt *btree) print1(depth int, offset uint64) {
	explan := ""
	if depth >= bt.treeLevels {
		explan += " LEAF"
	}
	print(strings.Repeat(" . ", depth)+"offset", offset, explan)
	nd := bt.getNode(offset)
	var sb strings.Builder
	sep := ""
	for it := nd.iter(); it.next(); {
		offset := it.offset
		if depth < bt.treeLevels {
			// tree
			print(strings.Repeat(" . ", depth)+strconv.Itoa(it.pos)+":",
				it.npre, it.diff, "=", it.known)
			bt.print1(depth+1, offset) // recurse
		} else {
			// leaf
			// print(strings.Repeat(" . ", depth)+strconv.Itoa(it.pos)+":",
			// 	strconv.Itoa(int(offset))+",", it.npre, it.diff, "=", it.known,
			// 	"("+bt.getLeafKey(offset)+")")
			sb.WriteString(sep)
			sep = " "
			if len(it.known) == 0 {
				sb.WriteString("''")
			} else {
				sb.Write(it.known)
			}
			// sb.WriteString(" = " + bt.getLeafKey(offset))
		}
	}
	if depth == bt.treeLevels {
		print(strings.Repeat(" . ", depth) + sb.String())
	}
}

//-------------------------------------------------------------------

func (bt *btree) StorSize() int {
	return 5 + 1
}

func (bt *btree) Write(w *stor.Writer) {
	w.Put5(bt.root).Put1(bt.treeLevels)
}

// ReadOverlay reads an Overlay from storage BUT without ixspec
func Read(st *stor.Stor, r *stor.Reader) *btree {
	root := r.Get5()
	treeLevels := r.Get1()
	return OpenBtree(st, root, treeLevels)
}

// trace ------------------------------------------------------------

const t = false // set to true to enable tracing

func trace(args ...interface{}) bool {
	fmt.Println(args...)
	return true
}
