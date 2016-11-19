package jnserver

func (c *client) hSet(key string, hkey string, value interface{}) {
	val, ok := c.base.Get(key)
	if ok {
		switch v := val.Value.(type) {
		case *map[string]string:
			(*v)[hkey] = value.(string)
			c.sendOk()
			return
		default:
			c.sendWrongType()
			return
		}
	}
	mp := make(map[string]string)
	mp[hkey] = value.(string)
	c.base.SetValue(key, &mp)
	c.sendOk()
}

func (c *client) hGet(key, hkey string) {
	val, ok := c.base.Get(key)
	if !ok {
		c.writer.SendNil()
	}

	switch mp := val.Value.(type) {
	case *map[string]string:
		v, ok := (*mp)[hkey]
		if ok {
			c.sendString(v)
			return
		}
		c.writer.SendNil()
		return
	default:
		c.sendWrongType()
		return
	}
}
