// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package dbms

import (
	"fmt"
	"hash/adler32"
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	"github.com/apmckinlay/gsuneido/db19"
	. "github.com/apmckinlay/gsuneido/runtime"
	"github.com/apmckinlay/gsuneido/util/str"
)

// DbmsLocal implements the Dbms interface using a local database
// i.e. standalone
type DbmsLocal struct {
	db        *db19.Database
	libraries []string //TODO concurrency
}

func NewDbmsLocal(db *db19.Database) IDbms {
	return &DbmsLocal{db: db}
}

// Dbms interface

var _ IDbms = (*DbmsLocal)(nil)

func (DbmsLocal) Admin(string) {
	panic("DbmsLocal Admin not implemented")
}

func (DbmsLocal) Auth(string) bool {
	panic("Auth only allowed on clients")
}

func (dbms DbmsLocal) Check() string {
	if err := dbms.db.Check(); err != nil {
		return fmt.Sprint(err)
	}
	return ""
}

func (DbmsLocal) Connections() Value {
	return EmptyObject
}

func (DbmsLocal) Cursor(string) ICursor {
	panic("DbmsLocal Cursor not implemented")
}

func (DbmsLocal) Cursors() int {
	panic("DbmsLocal Cursors not implemented")
}

func (dbms DbmsLocal) Dump(table string) string {
	var err error
	if table == "" {
		_, err = dbms.db.Dump("database.su")
	} else {
		_, err = dbms.db.DumpTable(table, table+".su")
	}
	if err != nil {
		return fmt.Sprint(err)
	}
	return ""
}

func (DbmsLocal) Exec(t *Thread, v Value) Value {
	fname := ToStr(ToContainer(v).ListGet(0))
	if i := strings.IndexByte(fname, '.'); i != -1 {
		ob := Global.GetName(t, fname[:i])
		m := fname[i+1:]
		return t.CallLookupEach1(ob, m, v)
	}
	fn := Global.GetName(t, fname)
	return t.CallEach1(fn, v)
}

func (DbmsLocal) Final() int {
	panic("DbmsLocal Final not implemented")
}

func (DbmsLocal) Get(int, string, Dir) (Row, *Header) {
	panic("DbmsLocal Get not implemented")
}

func (DbmsLocal) Info() Value {
	panic("DbmsLocal Info not implemented")
}

func (DbmsLocal) Kill(string) int {
	panic("DbmsLocal Kill not implemented")
}

func (DbmsLocal) Load(string) int {
	panic("DbmsLocal Load not implemented")
}

func (DbmsLocal) LibGet(name string) (result []string) {
	// Temporary version that reads from text files
	defer func() {
		if e := recover(); e != nil {
			panic("error loading " + name + " " + fmt.Sprint(e))
		}
	}()
	dir := "../stdlib/"
	hash := adler32.Checksum([]byte(name))
	name = strings.ReplaceAll(name, "?", "Q")
	file := dir + name + "_" + strconv.FormatUint(uint64(hash), 16)
	s, err := ioutil.ReadFile(file)
	if err != nil {
		if !strings.HasPrefix(name, "Rule_") {
			fmt.Println("LOAD", file, "NOT FOUND")
		}
		return nil
	}
	// fmt.Println("LOAD", name, "SUCCEEDED")
	return []string{"stdlib", string(s)}
}

func (DbmsLocal) Libraries() *SuObject {
	return NewSuObject()
}

func (DbmsLocal) Log(s string) {
	log.Println(s)
}

func (DbmsLocal) Nonce() string {
	panic("nonce only allowed on clients")
}

func (DbmsLocal) Run(string) Value {
	panic("DbmsLocal Run not implemented")
}

var sessionId string

func (DbmsLocal) SessionId(id string) string {
	if id != "" {
		sessionId = id
	}
	return sessionId
}

func (DbmsLocal) Size() int64 {
	panic("DbmsLocal Size not implemented")
}

func (DbmsLocal) Token() string {
	panic("DbmsLocal Token not implemented")
}

func (DbmsLocal) Transaction(bool) ITran {
	panic("DbmsLocal Transaction not implemented")
}

var prevTimestamp SuDate

func (DbmsLocal) Timestamp() SuDate {
	t := Now()
	if t.Equal(prevTimestamp) {
		t = t.Plus(0, 0, 0, 0, 0, 0, 1)
	}
	prevTimestamp = t
	return t
}

func (DbmsLocal) Transactions() *SuObject {
	panic("DbmsLocal Transactions not implemented")
}

func (dbms DbmsLocal) Unuse(lib string) bool {
	if lib == "stdlib" || !str.List(dbms.libraries).Has(lib) {
		return false
	}
	dbms.libraries = str.List(dbms.libraries).Without(lib)
	return true
}

func (dbms DbmsLocal) Use(lib string) bool {
	if str.List(dbms.libraries).Has(lib) {
		return false
	}
	//TODO check schema
	dbms.libraries = append(dbms.libraries, lib)
	return true
}

func (DbmsLocal) Close() {
}
