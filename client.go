package test_jn

import (
	"fmt"
	"net"

	"github.com/rpoletaev/test_jn/resp"
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
	c.reply(resp.FormatErrorFromString("Not Authenticated Use PASS your_password"))
}

func (c *Client) SendWrongPassword() {
	c.reply(resp.FormatErrorFromString("NotAuthenticated WrongPassword"))
}

func (c *Client) SendWrongType() {
	c.reply(resp.FormatErrorFromString("WRONGTYPE Operation against a key holding the wrong kind of value"))
}
func (c *Client) SendError(err string) {
	c.reply(resp.FormatErrorFromString(err))
}

func (c *Client) SendUnknownCommand(name string) {
	c.reply(resp.FormatErrorFromString("UnknownCommand USE CMDS to get command list"))
}

func (c *Client) SendWrongParamCount() {
	c.reply(resp.FormatErrorFromString("wrong count of argument"))
}

func (c *Client) SendOk() {
	c.reply(resp.FormatString("OK"))
}
