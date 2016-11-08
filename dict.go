package test_jn

import (
	"github.com/rpoletaev/test_jn/resp"
)

func (c *Client) hSet(key string, hkey string, value interface{}) {
	val, ok := c.base.Get(key)
	if ok {
		switch val.Value.(type) {
		case *map[string]string:
			(*(val.Value.(*map[string]string)))[hkey] = value.(string)
			c.SendOk()
			return
		default:
			c.SendWrongType()
			return
		}
	}
	mp := make(map[string]string)
	mp[hkey] = value.(string)
	c.base.SetValue(key, mp)
	c.SendOk()
}

func (c *Client) hGet(key, hkey string) {
	val, ok := c.base.Get(key)
	if !ok {
		c.reply(resp.FormatNill())
	}

	switch val.Value.(type) {
	case *map[string]string:
		val, ok := (*(val.Value.(*map[string]string)))[hkey]
		if ok {
			c.reply(resp.FormatBulkString(val))
			return
		}
		c.reply(resp.FormatNill())
		return
	default:
		c.SendWrongType()
		return
	}
}
