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
		return true
	}
}

func (c *Client) GetList(key string) (*[]interface{}, error) {
	item, ok := c.base.Get(key)
	if !ok {
		return nil, fmt.Errorf("Key is not found")
	}

	if !IsList(item.Value) {
		return nil, fmt.Errorf("Wrong type")
	}

	return item.Value.(*[]interface{}), nil
}

func (c *Client) ListLPush(key string, values []interface{}) {
	list, err := c.GetList(key)
	if err != nil {
		c.reply(format_standart_err(err.Error()))
		return
	}

	lpush(list, values)
	c.reply(format_ok())
}

func (c *Client) ListRPush(key string, values []interface{}) {
	list, err := c.GetList(key)
	if err != nil {
		c.reply(format_standart_err(err.Error()))
		return
	}

	rpush(list, values)
	c.reply(format_ok())
}

func (c *Client) ListLPop(key string) {
	list, err := c.GetList(key)
	if err != nil {
		c.reply(format_standart_err(err.Error()))
		return
	}

	value := lpop(list)
	c.reply(format_bulk_string(value))
	return
}

func (c *Client) ListRPop(key string) (interface{}, error) {
	list, err := c.GetList(key)
	if err != nil {
		return nil, err
	}

	value := rpop(list)
	return value, nil
}

func (c *Client) ListIndex(key string, i int) (interface{}, error) {
	list, err := c.GetList(key)
	if err != nil {
		return nil, err
	}

	if len(*list) >= i {
		return nil, fmt.Errorf("List has'nt index %d", i)
	}

	return (*list)[i], nil
}

func (c *Client) ListInsertAfter(key string, i int, value interface{}) error {
	list, err := c.GetList(key)
	if err != nil {
		return err
	}

	if len(*list) >= i {
		return fmt.Errorf("List has'nt index %d", i)
	}

	insertAfter(list, i, value)
	return nil
}

func (c *Client) ListRemove(key string, i int) error {
	list, err := c.GetList(key)
	if err != nil {
		return err
	}

	if len(*list) >= i {
		return fmt.Errorf("List has'nt index %d", i)
	}

	deleteIndex(list, i)
	return nil
}

func (c *Client) ListSetIndex(key string, i int, value interface{}) error {
	list, err := c.GetList(key)
	if err != nil {
		return err
	}

	if len(*list) >= i {
		return fmt.Errorf("List has'nt index %d", i)
	}

	(*list)[i] = value
	return nil
}
