package resp

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

type testRow struct {
	ParseString string
	Expecting   []interface{}
}

func TestParseArray(t *testing.T) {
	println("TestParseArray")
	testTable := []testRow{
		testRow{
			"*2\r\n$4\r\nLLEN\r\n$6\r\nmylist\r\n",
			[]interface{}{"LLEN", "mylist"},
		},
		testRow{
			"*2\r\n$4\r\nLLEN\r\n$7\r\nмайлист\r\n",
			[]interface{}{"LLEN", "майлист"},
		},
	}

	for _, tr := range testTable {
		println("parse ", tr.ParseString)
		ok, tail, cnt, err := isArray(strings.Split(tr.ParseString, "\r\n"))
		if !ok {
			t.Error("!ok expected ok")
		}

		if ok && err != nil {
			t.Error(err)
		}

		if ok {
			res, tl, parseErr := ParseArray(tail, cnt)
			if parseErr != nil {
				t.Errorf("%s tail %q res %q", parseErr, tl, res)
			}

			if fmt.Sprintf("%q", res) != fmt.Sprintf("%q", tr.Expecting) {
				t.Errorf("Expecting %q got %q\n", tr.Expecting, res)
			}
		}
	}
}

func TestParseRespString(t *testing.T) {
	println("TestParseRespString")
	testTable := []testRow{
		testRow{
			"-SomeError\r\n+odin\r\n:35\r\n*2\r\n$4\r\nLLEN\r\n$6\r\nmylist\r\n",
			[]interface{}{errors.New("SomeError"), "odin", int64(35), []interface{}{"LLEN", "mylist"}},
		},
	}

	for _, tr := range testTable {
		println("parse ", tr.ParseString)
		res, parseErr := ParseRespString(tr.ParseString)
		if parseErr != nil {
			t.Errorf("%s %q res %q", parseErr, tr.Expecting, res)
		}

		if fmt.Sprintf("%q", res) != fmt.Sprintf("%q", tr.Expecting) {
			t.Errorf("Expecting %q got %q\n", tr.Expecting, res)
		}
	}
}
