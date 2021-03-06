// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	. "github.com/apmckinlay/gsuneido/runtime"
	"github.com/apmckinlay/gsuneido/util/hacks"
)

const bufsize = 4000000

var inbuf [bufsize]byte
var outbuf [bufsize]byte
var intbuf = make([]byte, 4)

var nrecs = 0
var nums = 0
var obs = 0

// Convert changes database.su from old format to new format
func main() {
	fin, err := os.Open("database.su")
	ckerr(err)
	in := bufio.NewReader(fin)
	s, err := in.ReadString('\n')
	ckerr(err)
	hdr := "Suneido dump 1.0\n"
	if s != hdr {
		fmt.Println("ERROR: got:", s, " expected:", hdr)
		os.Exit(1)
	}
	fout, err := ioutil.TempFile(".", "suneido*.tmp")
	ckerr(err)
	out := bufio.NewWriter(fout)
	out.WriteString("Suneido dump 2\n")
	n := 0
	for { // each table
		schema, err := in.ReadString('\n')
		if err == io.EOF {
			break
		}
		ckerr(err)
		if !strings.HasPrefix(schema, "====== ") {
			panic("bad schema: " + schema)
		}
		// fmt.Print(schema)
		_, err = out.WriteString(schema)
		ckerr(err)
		convertTable(in, out)
		n++
	}
	out.Flush()
	fin.Close()
	tmpname := fout.Name()
	fout.Close()
	err = os.Remove("database.su.bak")
	if err != nil && !os.IsNotExist(err) {
		fmt.Println("ERROR: couldn't remove database.su.bak")
		fmt.Println(err)
	}
	err = os.Rename("database.su", "database.su.bak")
	if err != nil {
		fmt.Println("ERROR: couldn't rename database.su to database.su.bak")
		fmt.Println(err)
	}
	err = os.Rename(tmpname, "database.su")
	if err != nil {
		fmt.Println("ERROR: couldn't rename", tmpname, "to database.su")
		fmt.Println(err)
	}
	fmt.Println("converted", n, "tables,", nrecs, "records")
}

func convertTable(in *bufio.Reader, out *bufio.Writer) {
	for { // each record
		_, err := io.ReadFull(in, intbuf)
		if err == io.EOF {
			break
		}
		ckerr(err)
		size := int(binary.LittleEndian.Uint32(intbuf))
		if size == 0 {
			out.Write(intbuf)
			break
		}
		if size > cap(inbuf) {
			fmt.Println(size)
		}
		_, err = io.ReadFull(in, inbuf[:size])
		ckerr(err)
		convertRecord(inbuf[:size], out)
		nrecs++
	}
}

func convertRecord(b []byte, out *bufio.Writer) {
	s := hacks.BStoS(b)
	inrec := OldRec(s)
	var tb RecordBuilder
	n := inrec.Count()
	for i := 0; i < n; i++ { // for each value
		s := inrec.Get(i)
		if s != "" && (s[0] == PackPlus || s[0] == PackMinus) {
			dn := UnpackNumberOld(s)
			tb.Add(dn)
			nums++
		} else if s != "" && s[0] == PackObject {
			ob := UnpackObjectOld(s)
			tb.Add(ob)
			obs++
		} else if s != "" && s[0] == PackRecord {
			rec := UnpackRecordOld(s)
			tb.Add(rec)
			obs++
		} else {
			tb.AddRaw(s)
		}
	}
	outrec := string(tb.Build())
	binary.BigEndian.PutUint32(intbuf, uint32(len(outrec)))
	out.Write(intbuf)
	out.WriteString(outrec)
}

func ckerr(err error) {
	if err != nil {
		panic(err.Error())
	}
}
