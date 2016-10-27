package test_juno

import (
	"container/heap"
	"container/list"
	"fmt"
	"strconv"
	"sync"
	"time"
)

var mu sync.RWMutex

const (
	//NX parameter to command
	NX = "nx"
	//XX parameter to command
	XX = "xx"
	//EX parameter to command
	EX = "ex"

	//errors
	wrongParamType = "WrongParamType"
)

//default ExpTime is -1
type BaseItem struct {
	Value   interface{}
	ExpTime int64
}

type Base struct {
	Number       int
	items        map[string]BaseItem
	expiringKeys *ExpSignHeap
}

//Run creates instance of db and run job which cleans expired keys
func (base *Base) Run() error {
	base.items = make(map[string]BaseItem)
	base.expiringKeys = &ExpSignHeap{}
	heap.Init(base.expiringKeys)

	//find expired keys and remove they
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for tick := range ticker.C {
			for base.expiringKeys.Len() > 0 {
				if (*base.expiringKeys)[0].expiration <= tick.Unix() {
					heap.Remove(base.expiringKeys, 0)
					delete(base.items, (*base.expiringKeys)[0].key)
				} else {
					break
				}
			}
		}
	}()
	return nil
}

//Keys return all matched keys
func (base Base) Keys(patern string) []string {
	if len(base.items) == 0 {
		return nil
	}

	keys := make([]string, len(base.items))
	i := 0
	for k := range base.items {
		i++
		keys[i] = k
	}

	return keys
}

//SetExpire set expiration time to key and add key to expiration heap.
//return 1 if timeout was set and 0 if key does not exist
func (base *Base) SetExpire(key string, seconds int) int {
	val, ok := base.items[key]
	if !ok {
		return 0
	}

	if val.ExpTime > 0 {
		for i := 0; i < base.expiringKeys.Len(); i++ {
			if (*base.expiringKeys)[i].key == key {
				heap.Remove(base.expiringKeys, i)
				break
			}
		}
	}

	expTime := time.Now().Add(time.Duration(seconds) * time.Second).Unix()
	val.ExpTime = expTime
	sign := ExpiringSign{
		expiration: expTime,
		key:        key,
	}
	heap.Push(base.expiringKeys, sign)
	base.items[key] = val
	return 1
}

func (base *Base) SetValue(key string, val interface{}, params map[string]string) string {
	exp := int64(-1)
	if v, ok := params[EX]; ok {
		intExp, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return format_err("WRONGTYPE", "Expiration must be an integer")
		}

		exp = intExp
	}

	bi := BaseItem{val, exp}
	base.items[key] = bi
	return format_ok()
}

//GetTTL returns -2 if the key does not exist
//returns -1 if the key exists but has no associated expire
//or returns remaining seconds
func (base *Base) GetTTL(key string) string {
	item, ok := base.items[key]
	if !ok {
		return format_int(-2)
	}

	return format_int(item.ExpTime)
}

//Get type name of value item
//return error if wrong type
func GetTypeName(key string) (typeName string, err error) {

	switch val.(type) {
	case *list.List:
		return "List", nil
	case string:
		return "String", nil
	case map[string]interface{}:
		return "Dictionary", nil
	default:
		return "", fmt.Errorf("%s has wrong type")
	}
}
