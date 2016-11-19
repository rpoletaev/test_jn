package jnserver

import (
	"fmt"
	"net"

	"github.com/rpoletaev/respio"
)

type client struct {
	authorized bool
	base       *Base
	srv        *server
	con        net.Conn
	reader     *respio.RESPReader
	writer     *respio.RESPWriter
}

func (c *client) sendString(str string) {
	c.writer.SendBulkString(str)
}

func (c *client) sendSimpleString(str string) {
	c.writer.SendSimpleString(str)
}

func (c *client) sendOk() {
	c.writer.SendSimpleString("OK")
}

func (c *client) sendNotAuthenticated() {
	c.writer.SendError("Not Authenticated Use PASS your_password")
}

func (c *client) sendWrongPassword() {
	c.writer.SendError("NotAuthenticated WrongPassword")
}

func (c *client) sendWrongType() {
	c.writer.SendError("WRONGTYPE Operation against a key holding the wrong kind of value")
}
func (c *client) sendError(err string) {
	c.writer.SendError(err)
}

func (c *client) sendUnknownCommand(name string) {
	c.writer.SendError("UnknownCommand USE CMDS to get command list")
}

func (c *client) sendWrongParamCount() {
	c.writer.SendError("Wrong count of argument")
}

func (c *client) sendWrongParamType(exp string) {
	c.writer.SendError(fmt.Sprintf("Unexpected param type. Expecting %s\n", exp))
}
