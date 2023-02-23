// Copyright Suneido Software Corp. All rights reserved.
// Governed by the MIT license found in the LICENSE file.

package trace

import (
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"sync"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type what int

// var cur = ClientServer
var cur = what(0)

func Set(w int) int {
	prev := cur
	cur = what(w)
	return int(prev)
}

const (
	Functions what = 1 << iota
	Statements
	Opcodes
	Records
	Libraries
	SlowQuery
	Query
	Symbol
	AllIndex
	Table
	Select
	TempIndex
	QueryOpt

	Console
	LogFile

	ClientServer
	Exceptions
	Globals

	JoinOpt
	Dbms
)

func (w what) String() string {
	return map[what]string{
		Functions:    "FUNC ",
		Statements:   "STMT ",
		Opcodes:      "OP ",
		Records:      "REC ",
		Libraries:    "LIB ",
		SlowQuery:    "SLOWQ ",
		Query:        "QUERY ",
		Symbol:       "SYM ",
		AllIndex:     "ALLIDX ",
		Table:        "TABLE ",
		Select:       "SELECT ",
		TempIndex:    "TMPIDX ",
		QueryOpt:     "QOPT ",
		Console:      "CONSOLE ",
		LogFile:      "LOGFILE ",
		ClientServer: "CS ",
		Exceptions:   "EXCEPT ",
		Globals:      "GLOBAL ",
		JoinOpt:      "JOINOPT ",
		Dbms:         "DBMS ",
	}[w]
}

func (w what) Set() {
	cur |= w
}

func (w what) Println(first any, rest ...any) {
	// kept short in hopes it will be inlined
	if cur&w != 0 {
		format(&first)
		for i := range rest {
			format(&rest[i])
		}
		s := w.String() + fmt.Sprint(first) + " " + fmt.Sprintln(rest...)
		Print(s)
	}
}

func Println(args ...any) {
	for i := range args {
		format(&args[i])
	}
	Print(fmt.Sprintln(args...))
}

func Print(s string) {
	if cur&LogFile != 0 || cur&(LogFile|Console) == 0 {
		logPrint(s)
	}
	if cur&Console != 0 || cur&(LogFile|Console) == 0 {
		consolePrint(s)
	}
}

type writer struct{}

func (w writer) Write(p []byte) (n int, err error) {
	Print(string(p))
	return len(p), nil
}

func (w writer) WriteString(s string) (n int, err error) {
	Print(s)
	return len(s), nil
}

var Writer writer

var printer = message.NewPrinter(language.English)

func format(p *any) {
	switch x := (*p).(type) {
	case []string:
		if x == nil {
			*p = "nil"
		}
	case int, uint, int32, uint32, int64, uint64:
		// add commas to make big numbers more readable
		*p = Number(*p)
	case float32, float64:
		*p = fmt.Sprintf("%.4g", *p)
	}
}

const impossible = math.MaxInt / 64 // must match query

func Number(n any) string {
	if n, ok := n.(int); ok && n >= impossible {
		return "impossible"
	}
	return strings.Replace(printer.Sprintf("%d", n), ",", "_", -1)
}

func (w what) On() bool {
	return cur&w != 0
}

var traceLog *os.File
var traceLogOnce sync.Once

func logPrint(s string) {
	traceLogOnce.Do(func() {
		var err error
		traceLog, err = os.OpenFile("trace.log",
			os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Println("ERROR", err)
		}
	})
	if traceLog != nil {
		traceLog.WriteString(s)
	}
}
