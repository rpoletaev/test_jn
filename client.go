package test_jn

import (
	"fmt"
	"net"
)

type Client struct {
	authorized bool
	base       *Base
	srv        *server
	con        net.Conn
}

func (cli *Client) reply(str string) {
	fmt.Fprint(cli.con, str)
}

func (c *Client) SendNotAuthenticated() {
	err := format_err("Not Authenticated", "Use PASS your_password")
	c.reply(err)
}

func (c *Client) SendWrongPassword() {
	c.reply(format_err("NotAuthenticated", "WrongPassword"))
}

func (c *Client) SendError(err string) {
	c.reply(format_standart_err(err))
}

func (c *Client) SendUnknownCommand(name string) {
	err := format_err("UnknownCommand USE", "CMDS to get command list")
	c.reply(err)
}

func (c *Client) SendWrongParamCount() {
	c.reply(format_standart_err("wrong count of argument"))
}
func (c *Client) SendOk() {
	c.reply(format_ok())
}

func (c *Client) SendVal(val fmt.Stringer) {
	c.reply(val.String())
}

func (c *Client) SendString(str string) {
	c.reply(format_str(str))
}

func (c *Client) SendNil() {
	c.reply(format_nill())
}

func (c *Client) SendBulk(str string) {
	c.reply(format_bulk_string(str))
}

func (c *Client) SendInt(n int64) {
	c.reply(format_int(n))
}
