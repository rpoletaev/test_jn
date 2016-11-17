package test_jn

import "strconv"
import "fmt"

type handler func(client *client, params ...interface{})

func passCommand(c *client, prs ...interface{}) {
	if len(prs) < 1 {
		c.sendWrongParamCount()
		return
	}

	if c.srv.Password != string(prs[0].([]byte)) {
		c.sendWrongPassword()
		return
	}

	c.srv.clients[c] = true
	c.sendOk()
}

//strings
func getCommand(c *client, prs ...interface{}) {
	if len(prs) < 1 {
		c.sendError("Wrong param count")
	}

	key, err := getStringFromParam(prs[0])
	if err != nil {
		c.sendError(err.Error())
		return
	}

	val, exist := c.base.Get(key)
	if exist {
		switch v := val.Value.(type) {
		case []byte:
			c.writer.SendBulk(v)
			c.writer.Flush()
			return
		default:
			fmt.Printf("Value has wrong type %T\n", v)
			c.sendError("Wrong value type")
			return
		}
	}

	c.writer.SendNil()
}

func setCommand(c *client, prs ...interface{}) {
	if len(prs) < 2 {
		c.sendWrongParamCount()
		return
	}

	key, err := getStringFromParam(prs[0])
	if err != nil {
		c.sendError(err.Error())
		return
	}

	if len(prs) > 2 {
		switch prs[2] {
		case EX:
			if len(prs) < 4 {
				c.sendWrongParamCount()
				return
			}

			ttl, err := strconv.ParseInt(prs[3].(string), 10, 64)
			if err != nil {
				c.sendError("TTL must be an integer")
				return
			}
			c.base.SetValueWithTTL(key, prs[1], ttl)
			c.sendOk()
			return
		case NX:
			if c.base.SetIfNotExist(key, prs[1]) {
				c.writer.SendRESPInt(1)
			} else {
				c.writer.SendRESPInt(0)
			}
			return
		case XX:
			if c.base.SetIfExist(key, prs[1]) {
				c.writer.SendRESPInt(1)
			} else {
				c.writer.SendRESPInt(0)
			}
			return
		}
	}

	c.base.SetValue(key, prs[1])
	c.sendOk()
}

func keysCommand(c *client, patern ...interface{}) {
	if len(patern) == 0 {
		keys := c.base.AllKeys()
		c.writer.SendArray(int64(len(keys)))
		for _, str := range keys {
			c.sendString(str)
		}
		c.writer.Flush()
		return
	}

	ptrn, err := getStringFromParam(patern[0])
	if err != nil {
		c.sendError(err.Error())
		return
	}

	keys, err := c.base.Keys(ptrn)
	if err != nil {
		c.sendError(err.Error())
		return
	}

	keys = c.base.AllKeys()
	c.writer.SendArray(int64(len(keys)))
	for _, str := range keys {
		c.sendString(str)
	}
	c.writer.Flush()
	return
}

func delCommand(c *client, keys ...interface{}) {
	for _, k := range keys {
		c.base.Remove(string(k.([]byte)))
	}

	c.writer.SendRESPInt(int64(len(keys)))
	c.writer.Flush()
}

func getStringFromParam(i interface{}) (string, error) {
	switch v := i.(type) {
	case []byte:
		return string(v), nil
	default:
		return emptyString, fmt.Errorf("Wrong type '%T' of key", v)
	}
}

const emptyString = ""
