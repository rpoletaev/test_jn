package test_jn

import (
	"fmt"
	"log"
	"strconv"

	"github.com/rpoletaev/test_jn/resp"
)

type handler func(client *Client, params ...interface{})

func passCommand(c *Client, prs ...interface{}) {
	if len(prs) < 1 {
		c.SendWrongParamCount()
		return
	}

	if c.srv.Password != prs[0].(string) {
		c.SendWrongPassword()
		return
	}

	c.srv.clients[c] = true
	c.SendOk()
}

//strings
func getCommand(c *Client, prs ...interface{}) {
	if len(prs) < 1 {
		c.reply(resp.FormatError(fmt.Errorf("Wrong param count")))
	}
	val, exist := c.base.Get(prs[0].(string))
	if exist {
		switch val.Value.(type) {
		case string:
			c.reply(resp.FormatString(val.Value.(string)))
			return
		default:
			c.reply(resp.FormatErrorFromString("Wrong value type"))
			log.Printf("%v", val)
			return
		}
	}

	c.reply(resp.FormatNill())
}

func setCommand(c *Client, prs ...interface{}) {
	if len(prs) < 2 {
		c.SendWrongParamCount()
		return
	}

	if len(prs) > 2 {
		switch prs[2] {
		case EX:
			if len(prs) < 4 {
				c.SendWrongParamCount()
				return
			}

			ttl, err := strconv.ParseInt(prs[3].(string), 10, 64)
			if err != nil {
				c.SendError("TTL must be an integer")
				return
			}
			c.base.SetValueWithTTL(prs[0].(string), prs[1], ttl)
			c.SendOk()
			return
		case NX:
			if c.base.SetIfNotExist(prs[0].(string), prs[1]) {
				c.reply(resp.FormatInt(1))
			} else {
				c.reply(resp.FormatInt(0))
			}
			return
		case XX:
			if c.base.SetIfExist(prs[0].(string), prs[1]) {
				c.reply(resp.FormatInt(1))
			} else {
				c.reply(resp.FormatInt(0))
			}
			return
		}
	}

	c.base.SetValue(prs[0].(string), prs[1])
	c.SendOk()
}

func keysCommand(c *Client, patern ...interface{}) {
	if len(patern) == 0 {
		keys := c.base.AllKeys()
		respArr := resp.FormatBulkStringArray(keys)
		c.reply(string(respArr))
		return
	}

	keys, err := c.base.Keys(patern[0].(string))
	if err != nil {
		c.SendError(err.Error())
		return
	}

	respArr := resp.FormatBulkStringArray(keys)
	c.reply(string(respArr))
}

func delCommand(c *Client, keys ...interface{}) {
	res := c.base.Remove(keys...)
	c.reply(resp.FormatInt(res))
}
