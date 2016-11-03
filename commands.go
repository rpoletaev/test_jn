package test_jn

import (
	"log"
	"strconv"
)

type handler func(client *Client, params ...string)

func passCommand(c *Client, prs ...string) {
	if len(prs) < 1 {
		c.SendWrongParamCount()
		return
	}

	if c.srv.Password != prs[0] {
		c.SendWrongPassword()
		return
	}

	c.srv.clients[c] = true
	c.SendOk()
}

//strings
func getCommand(c *Client, prs ...string) {
	if len(prs) < 1 {
		c.SendWrongParamCount()
		return
	}

	val, exist := c.base.Get(prs[0])
	if exist {
		switch str := val.Value.(type) {
		case string:
			log.Println("PRINT STRING")
			c.SendString(str)
			return
		default:
			c.SendError("wrong type")
			log.Printf("%v", val)
			return
		}
	}

	c.SendNil()
}

func setCommand(c *Client, prs ...string) {
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

			ttl, err := strconv.ParseInt(prs[3], 10, 64)
			if err != nil {
				c.SendError("TTL must be an integer")
				return
			}
			c.base.SetValueWithTTL(prs[0], prs[1], ttl)
			c.SendOk()
			return
		case NX:
			if c.base.SetIfNotExist(prs[0], prs[1]) {
				c.SendInt(1)
			} else {
				c.SendInt(0)
			}
			return
		case XX:
			if c.base.SetIfExist(prs[0], prs[1]) {
				c.SendInt(1)
			} else {
				c.SendInt(0)
			}
			return
		}
	}

	c.base.SetValue(prs[0], prs[1])
	c.SendOk()
}

func keysCommand(c *Client, patern ...string) {
	if len(patern) == 0 {
		keys := c.base.AllKeys()
		c.reply(format_bulk_string(keys...))
		return
	}

	keys, err := c.base.Keys(patern[0])
	if err != nil {
		c.reply(format_standart_err(err.Error()))
		return
	}

	c.reply(format_bulk_string(keys...))
}

func delCommand(c *Client, keys ...string) {
	res := c.base.Remove(keys...)
	c.reply(format_int(res))
}
