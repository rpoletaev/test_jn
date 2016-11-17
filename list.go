package test_jn

import (
	"fmt"
)

func lpush(l *[]interface{}, newList []interface{}) {
	*l = append(newList, *l...)
}

func rpush(l *[]interface{}, newList []interface{}) {
	*l = append(*l, newList...)
}

func lpop(l *[]interface{}) interface{} {
	item := (*l)[0]
	newList := make([]interface{}, len(*l)-1)
	copy(newList, (*l)[1:])
	*l = newList
	return item
}

func rpop(l *[]interface{}) interface{} {
	item := (*l)[len(*l)-1]
	*l = (*l)[0 : len(*l)-1]
	return item
}

func insertAfter(l *[]interface{}, i int, value interface{}) {
	newList := make([]interface{}, len(*l)+1)
	copy(newList[:i+1], (*l)[:i+1])
	copy(newList[i+2:], (*l)[i+1:])
	newList[i+1] = value
	*l = newList
}

func deleteIndex(l *[]interface{}, i int) {
	copy((*l)[i:], (*l)[i+1:])
	*l = (*l)[:len((*l))-1]
}

func IsList(value interface{}) bool {
	switch value.(type) {
	case *[]interface{}:
		return true
	default:
		return false
	}
}

func (c *client) GetList(key string) (*[]interface{}, error) {
	item, ok := c.base.Get(key)
	if !ok {
		return nil, nil
	}

	if !IsList(item.Value) {
		return nil, fmt.Errorf("Wrong type")
	}

	return item.Value.(*[]interface{}), nil
}

func (c *client) ListLPush(key string, values []interface{}) {
	list, err := c.GetList(key)
	if err != nil {
		c.sendError(err.Error())
		return
	}

	if list == nil {
		lst := make([]interface{}, len(values))
		copy(lst, values)
		c.base.SetValue(key, &lst)
		c.sendOk()
		return
	}

	lpush(list, values)
	c.sendOk()
}

func (c *client) ListRPush(key string, values []interface{}) {
	list, err := c.GetList(key)
	if err != nil {
		c.sendError(err.Error())
		return
	}

	if list == nil {
		lst := make([]interface{}, len(values))
		copy(lst, values)
		c.base.SetValue(key, &lst)
		c.sendOk()
		return
	}

	rpush(list, values)
	c.sendOk()
}

func (c *client) ListLPop(key string) {
	list, err := c.GetList(key)
	if err != nil {
		c.sendError(err.Error())
		return
	}

	value := lpop(list)
	c.writer.SendBulk(value.([]byte))
	c.writer.Flush()
}

func (c *client) ListRPop(key string) {
	list, err := c.GetList(key)
	if err != nil {
		c.sendError(err.Error())
		return
	}

	value := rpop(list)
	c.writer.SendBulk(value.([]byte))
	c.writer.Flush()
}

func (c *client) ListIndex(key string, i int) {
	list, err := c.GetList(key)
	if err != nil {
		c.sendError(err.Error())
		return
	}

	if i < 0 {
		c.sendError("Index should be greather than 0")
		return
	}

	if len(*list) <= i {
		fmt.Printf("List '%s': %v", key, *list)
		c.sendError(fmt.Sprintf("list has lenght %d. List has'nt index %d\n", len(*list), i))
		return
	}

	c.writer.SendBulk((*list)[i].([]byte))
	c.writer.Flush()
}

func (c *client) ListInsertAfter(key string, i int, value interface{}) {
	list, err := c.GetList(key)
	if err != nil {
		c.sendError(err.Error())
		return
	}

	if i < 0 {
		c.sendError("Index should be greather than 0")
		return
	}

	if len(*list) <= i {
		c.sendError(fmt.Sprintf("List has'nt index %d", i))
		return
	}

	insertAfter(list, i, value)
	c.sendOk()
}

func (c *client) ListRemove(key string, i int) {
	list, err := c.GetList(key)
	if err != nil {
		c.sendError(err.Error())
	}

	if i < 0 {
		c.sendError("Index should be greather than 0")
		return
	}

	if len(*list) <= i {
		c.sendError(fmt.Sprintf("List has'nt index %d", i))
		return
	}

	deleteIndex(list, i)
	c.sendOk()
}

func (c *client) ListSetIndex(key string, i int, value interface{}) {
	list, err := c.GetList(key)
	if err != nil {
		c.sendError(err.Error())
		return
	}

	if i < 0 {
		c.sendError("Index should be greather than 0")
		return
	}

	if len(*list) >= i {
		c.sendError(fmt.Sprintf("List has'nt index %d", i))
		return
	}

	(*list)[i] = value
	c.sendOk()
}

func (c *client) ListLength(key string) {
	list, err := c.GetList(key)
	if err != nil {
		c.sendError(err.Error())
		return
	}

	c.writer.SendRESPInt(int64(len(*list)))
	c.writer.Flush()
}
