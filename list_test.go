package test_jn

import (
	"fmt"
	"testing"
)

func TestListLPush(t *testing.T) {
	println("TestListPush")
	var l []interface{}
	expl := []interface{}{1, 2, 3}
	lpush(&l, []interface{}{1, 2, 3})
	if l[0] != expl[0] || l[1] != expl[1] || l[2] != expl[2] {
		t.Error("Expected ", expl, "got ", l)
	}

	// lpush(&l, 4, 5, 6)
	func(oldlst *[]interface{}, newlst []interface{}) {
		lpush(oldlst, newlst)
	}(&l, []interface{}{4, 5, 6})
	expl = []interface{}{4, 5, 6, 1, 2, 3}
	if fmt.Sprintf("%q", l) != fmt.Sprintf("%q", expl) {
		t.Error("Expected ", expl, "got ", l)
	}
}

func TestListRPush(t *testing.T) {
	println("TestListRPush")
	var l []interface{}
	expl := []interface{}{1, 2, 3}
	rpush(&l, []interface{}{1, 2, 3})
	if l[0] != expl[0] || l[1] != expl[1] || l[2] != expl[2] {
		t.Error("Expected ", expl, "got ", l)
	}

	expl = []interface{}{1, 2, 3, 4, 5, 6}
	rpush(&l, []interface{}{4, 5, 6})
	if fmt.Sprintf("%q", l) != fmt.Sprintf("%q", expl) {
		t.Error("Expected ", expl, "got ", l)
	}
}

func TestListLPop(t *testing.T) {
	println("TestListLPop")
	l := &[]interface{}{1, 2, 3}
	val := lpop(l)
	if fmt.Sprintf("%d", val) != fmt.Sprintf("%d", 1) {
		t.Error("Expected ", 1, " got ", val)
	}

	if fmt.Sprintf("%q", *l) != fmt.Sprintf("%q", []interface{}{2, 3}) {
		t.Error("Expected ", []interface{}{2, 3}, " got ", *l)
	}
}

func TestListRPop(t *testing.T) {
	println("TestListRPop")
	l := &[]interface{}{1, 2, 3}
	val := rpop(l)
	if fmt.Sprintf("%d", val) != fmt.Sprintf("%d", 3) {
		t.Error("Expected ", 3, " got ", val)
	}

	if fmt.Sprintf("%q", *l) != fmt.Sprintf("%q", []interface{}{1, 2}) {
		t.Errorf("Expected %q got %q\n", []interface{}{1, 2}, *l)
	}
}

func TestListInsertAfter(t *testing.T) {
	println("TestListInsertAfter")
	l := []interface{}{1, 2, 4, 5}
	insertAfter(&l, 1, 3)
	if fmt.Sprintf("%q", l) != fmt.Sprintf("%q", []interface{}{1, 2, 3, 4, 5}) {
		t.Errorf("Expected %q got %q\n", []interface{}{1, 2, 3, 4, 5}, l)
	}
}

func TestListDeleteIndex(t *testing.T) {
	println("TestListDeleteIndex")
	l := &[]interface{}{1, 2, 3, 4}
	deleteIndex(l, 1)
	if fmt.Sprintf("%q", *l) != fmt.Sprintf("%q", []interface{}{1, 3, 4}) {
		t.Errorf("Expected %q got %q\n", []interface{}{1, 3, 4}, *l)
	}
}
