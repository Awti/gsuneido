package runtime

import (
	"github.com/apmckinlay/gsuneido/util/pack"
	"github.com/apmckinlay/gsuneido/util/verify"
	"strings"
)

/*
Record is an immutable record stored in a string
using the same format as cSuneido and jSuneido.

NOTE: This is the post 2019 format using a two byte header.

It is used for storing data records in the database
and for transferring data records across the client-server protocol.

An empty Record is a single zero byte.

First two bytes are the type and the count of values, high two bits are the type
followed by the total length (uint8, uint16, or uint32)
followed by the offsets of the fields (uint8, uint16, or uint32)
followed by the contents of the fields
integers are stored big endian (most significant first)
*/
type Record string

const (
	type8 = iota + 1
	type16
	type32
)
const sizeMask = 0x3ff

const hdrlen = 2

// Count returns the number of values in the record
func (r Record) Count() int {
	if r[0] == 0 {
		return 0
	}
	return (int(r[0])<<8 + int(r[1])) & sizeMask
}

// GetVal is a convenience method to get and unpack
func (r Record) GetVal(i int) Value {
	return Unpack(r.GetRaw(i))
}

// Get returns one of the (usually packed) values
func (r Record) GetRaw(i int) string {
	if i >= r.Count() {
		return ""
	}
	var pos, end int
	switch r.mode() {
	case type8:
		j := hdrlen + i
		end = int(r[j])
		pos = int(r[j+1])
	case type16:
		j := hdrlen + 2*i
		end = (int(r[j]) << 8) | int(r[j+1])
		pos = (int(r[j+2]) << 8) | int(r[j+3])
	case type32:
		j := hdrlen + 4*i
		end = (int(r[j]) << 24) | (int(r[j+1]) << 16) |
			(int(r[j+2]) << 8) | int(r[j+3])
		pos = (int(r[j+4]) << 24) | (int(r[j+5]) << 16) |
			(int(r[j+6]) << 8) | int(r[j+7])
	default:
		panic("invalid record type")
	}
	return string(r)[pos:end]
}

func (r Record) mode() byte {
	return r[0] >> 6
}

func (r Record) String() string {
	var sb strings.Builder
	sep := "<"
	for i := 0; i < r.Count(); i++ {
		sb.WriteString(sep)
		sep = ", "
		sb.WriteString(r.GetVal(i).String())
	}
	sb.WriteString(">")
	return sb.String()
}

// ------------------------------------------------------------------

// RecordBuilder is used to construct records. Zero value is ready to use.
type RecordBuilder struct {
	vals []Packable
}

const MaxValues = 0x3fff

// Add appends a Packable
func (b *RecordBuilder) Add(p Packable) *RecordBuilder {
	b.vals = append(b.vals, p)
	return b
}

// AddRaw appends a string containing an already packed value
func (b *RecordBuilder) AddRaw(s string) *RecordBuilder {
	b.Add(Packed(s))
	return b
}

// Packed is a Packable wrapper for an already packed value
type Packed string

func (p Packed) Pack(buf *pack.Encoder) {
	buf.PutStr(string(p))
}

func (p Packed) PackSize(int) int {
	return len(p)
}

// Build

func (b *RecordBuilder) Build() Record {
	if len(b.vals) > MaxValues {
		panic("too many values for record")
	}
	if len(b.vals) == 0 {
		return Record("\x00")
	}
	sizes := make([]int, len(b.vals))
	for i, v := range b.vals {
		sizes[i] = v.PackSize(0)
	}
	length := b.recSize(sizes)
	buf := pack.NewEncoder(length)
	b.build(buf, length, sizes)
	verify.That(len(buf.String()) == length) //TODO remove
	return Record(buf.String())
}

func (b *RecordBuilder) recSize(sizes []int) int {
	nfields := len(b.vals)
	datasize := 0
	for _, size := range sizes {
		datasize += size
	}
	return tblength(nfields, datasize)
}

func tblength(nfields, datasize int) int {
	if nfields == 0 {
		return 1
	}
	length := hdrlen + (1 + nfields) + datasize
	if length < 0x100 {
		return length
	}
	length = hdrlen + 2*(1+nfields) + datasize
	if length < 0x10000 {
		return length
	}
	return hdrlen + 4*(1+nfields) + datasize
}

func (b *RecordBuilder) build(dst *pack.Encoder, length int, sizes []int) {
	b.buildHeader(dst, length, sizes)
	nfields := len(b.vals)
	for i := nfields - 1; i >= 0; i-- {
		b.vals[i].Pack(dst)
	}
}

func (b *RecordBuilder) buildHeader(dst *pack.Encoder, length int, sizes []int) {
	mode := mode(length)
	nfields := len(b.vals)
	dst.Uint16(uint16(mode<<14 | nfields))
	b.buildOffsets(dst, length, sizes)
}

func (b *RecordBuilder) buildOffsets(dst *pack.Encoder, length int, sizes []int) {
	nfields := len(b.vals)
	offset := length
	switch mode(length) {
	case type8:
		dst.Put1(byte(offset))
		for i := 0; i < nfields; i++ {
			offset -= sizes[i]
			dst.Put1(byte(offset))
		}
	case type16:
		dst.Uint16(uint16(offset))
		for i := 0; i < nfields; i++ {
			offset -= sizes[i]
			dst.Uint16(uint16(offset))
		}
	case type32:
		dst.Uint32(uint32(offset))
		for i := 0; i < nfields; i++ {
			offset -= sizes[i]
			dst.Uint32(uint32(offset))
		}
	}
}

func mode(length int) int {
	if length == 0 {
		return 0
	} else if length < 0x100 {
		return type8
	} else if length < 0x10000 {
		return type16
	} else {
		return type32
	}
}
