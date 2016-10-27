package test_juno

import "net"

type Client struct {
	authorized bool
	basenumber int
	//srv        *server
	con net.Conn
}

func (cli *Client) reply(str string) string {
	if _, err := cli.con.Write([]byte(str)); err != nil {
		return err
	}

	return nil
}
