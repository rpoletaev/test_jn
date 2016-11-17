package test_jn

import (
	"fmt"
	"log"
	"regexp"
	"sync"
	"time"

	"github.com/rpoletaev/test_jn/ttl"
)

const (
	//NX parameter to set if not exist
	NX = "nx"
	//XX parameter to command
	XX = "xx"
	//EX parameter to set expiration in SET command
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
	Number int
	sync.RWMutex
	items        map[string]BaseItem
	expiringKeys *ttl.ExpiringQueue
}

//Run creates instance of db and run job which cleans expired keys
func (base *Base) Run() error {
	base.items = make(map[string]BaseItem)
	base.expiringKeys = ttl.CreateExpiringQueue(&ttl.ExpQueueOptions{
		Period: time.Duration(5) * time.Second,
		DoIfExpired: func(key string) {
			base.Remove(key)
		},
		DoIfStoped: func() {
			log.Println("Queue stopped")
		},
	})
	return nil
}

func (base *Base) Stop() {
	base.expiringKeys.Stop()
	log.Println("Base ", base.Number, "is stopped")
}

//Keys return all matched keys
func (base Base) Keys(patern string) ([]string, error) {
	if base.Len() == 0 {
		return nil, nil
	}

	allKeys := base.AllKeys()
	if patern == "" {
		return allKeys, nil
	}

	re, err := regexp.Compile(patern)
	if err != nil {
		return nil, err
	}

	keys := make([]string, len(allKeys)/2)
	for _, key := range allKeys {
		if re.MatchString(key) {
			keys = append(keys, key)
		}
	}

	return keys, nil
}

//AllKeys return all keys from base
func (base *Base) AllKeys() []string {
	if base.Len() == 0 {
		return nil
	}

	keys := make([]string, len(base.items))
	i := 0
	base.RLock()
	defer base.RUnlock()
	for k := range base.items {
		keys[i] = k
		i++
	}
	return keys
}

func (base *Base) Len() int {
	base.RLock()
	defer base.RUnlock()
	return len(base.items)
}

// Get checks contains base key and return Item with true values
// if key is expired then delete key from base and return empty Item and false
func (base *Base) Get(key string) (BaseItem, bool) {
	base.RLock()
	val, ok := base.items[key]
	base.RUnlock()
	if val.ExpTime >= 0 && val.ExpTime <= time.Now().Unix() {
		base.Remove(key)
		return BaseItem{}, false
	}
	return val, ok
}

func (base *Base) Set(key string, bi BaseItem) {
	base.Lock()
	base.items[key] = bi
	base.Unlock()
}

func (base *Base) SetIfNotExist(key string, value interface{}) bool {
	_, exist := base.Get(key)
	if exist {
		return false
	}

	base.SetValue(key, value)
	return true
}

func (base *Base) SetIfExist(key string, value interface{}) bool {
	bi, exist := base.Get(key)
	if !exist {
		return false
	}

	bi.Value = value
	base.Set(key, bi)
	return true
}

//SetValue Set value to key
func (base *Base) SetValue(key string, val interface{}) {
	exp := int64(-1)
	bi := BaseItem{val, exp}
	base.Set(key, bi)
}

//SetValueWithTTL Set value to key then set ttl to key
func (base *Base) SetValueWithTTL(key string, val interface{}, ttl int64) {
	base.SetValue(key, val)
	base.SetTTL(key, ttl)
}

//SetExpire set expiration time to key and add key to expiration heap.
//return 1 if timeout was set and 0 if key does not exist
func (base *Base) SetTTL(key string, seconds int64) int64 {
	val, ok := base.Get(key)
	if !ok {
		return 0
	}

	expTime := time.Now().Add(time.Duration(seconds) * time.Second).Unix()
	val.ExpTime = expTime
	sign := ttl.ExpiringSign{
		Expiration: expTime,
		Key:        key,
	}
	base.expiringKeys.Push(sign)
	base.Set(key, val)
	return 1
}

//GetTTL returns -2 if the key does not exist
//returns -1 if the key exists but has no associated expire
//or returns remaining seconds
func (base *Base) GetTTL(key string) int64 {
	if item, ok := base.Get(key); ok {
		return item.ExpTime
	}

	return -2
}

//Get type name of value item
//return error if wrong type
func (base *Base) GetTypeName(key string) (typeName string, err error) {
	item, ok := base.Get(key)
	if !ok {
		return "", fmt.Errorf("key %s is not found", key)
	}
	switch item.Value.(type) {
	case *[]interface{}:
		return "List", nil
	case string:
		return "String", nil
	case map[string]interface{}:
		return "Dictionary", nil
	default:
		return "", fmt.Errorf("%s has wrong type", key)
	}
}

//Remove remove item from base
// returns count of removing keys
func (base *Base) Remove(key string) {
	base.Lock()
	delete(base.items, key)
	base.Unlock()
}
