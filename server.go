package test_jn

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/rpoletaev/test_jn/resp"
	"github.com/xlab/closer"
)

const (
	NotAuthorized = "User not authorized"
)

type ServerConfig struct {
	Port        int    `yaml:"port"`
	Requirepass bool   `yaml:"requirepass"`
	Password    string `yaml:"password"`
}

type command struct {
	authRequired bool
	execute      handler
	writeToAof   bool
}

type server struct {
	*ServerConfig
	clients  map[*Client]bool
	bases    []*Base
	commands map[string]command
}

//CreateServer Creates server with default config
func CreateServer() *server {
	defaultCfg := &ServerConfig{
		Port:        2020,
		Requirepass: true,
		Password:    "roma",
	}

	return CreateServerWithConfig(defaultCfg)
}

//CreateServerWithConfig Creates server with user config
func CreateServerWithConfig(c *ServerConfig) *server {
	return &server{
		c,
		make(map[*Client]bool),
		nil,
		make(map[string]command),
	}

}

func (server *server) Run() {
	defer closer.Close()
	if err := server.initBases(); err != nil {
		log.Fatal(err)
	}

	server.loadCommands()

	l, err := net.Listen("tcp", fmt.Sprint(":", server.Port))
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	log.Println("Listen on ", server.Port, "...")

	for {
		con, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go func(nc net.Conn) {
			defer func() {
				nc.Close()
			}()

			input := bufio.NewReader(nc)
			cli := &Client{
				authorized: !server.Requirepass,
				base:       server.bases[0],
				srv:        server,
				con:        nc,
			}

			server.clients[cli] = !server.Requirepass
			log.Println("client connected")
			for {
				cmdName, prs, err := server.parseCommand(input)
				if err != nil {
					cli.reply(resp.FormatError(err))
				}
				cmd, exist := server.commands[cmdName]
				if !exist {
					cli.SendUnknownCommand(cmdName)
				} else {
					if authenticated, ok := server.clients[cli]; ok && !authenticated && cmd.authRequired {
						cli.SendNotAuthenticated()
					} else {
						cmd.execute(cli, prs...)
					}
				}
			}
		}(con)
	}
}

func (server *server) Stop() {
	log.Println("stopping server")
	for _, b := range server.bases {
		b.Stop()
	}
}
func (s *server) initBases() error {
	s.bases = make([]*Base, 0)
	s.newBase()
	return nil
}

func (s *server) newBase() (baseNum int) {
	baseNum = len(s.bases)
	newBase := &Base{Number: baseNum}
	newBase.Run()
	s.bases = append(s.bases, newBase)
	return baseNum
}

func (s *server) loadCommands() {
	// SERVER COMMANDS
	s.commands["PASS"] = command{
		false,
		passCommand,
		false,
	}

	s.commands["CMDS"] = command{
		false,
		func(cli *Client, prs ...interface{}) {
			cmds := make([]string, len(s.commands))
			var counter int
			for kc := range s.commands {
				cmds[counter] = kc
				counter++
			}
			respArr := resp.FormatBulkStringArray(cmds)
			cli.reply(string(respArr))
		},
		false,
	}

	// STRING COMMANDS
	s.commands["GET"] = command{
		true,
		getCommand,
		false,
	}

	s.commands["SET"] = command{
		true,
		setCommand,
		false,
	}

	s.commands["KEYS"] = command{
		true,
		keysCommand,
		false,
	}

	s.commands["DEL"] = command{
		true,
		delCommand,
		true,
	}

	//LIST COMMANDS
	s.commands["LPUSH"] = command{
		true,
		func(cli *Client, prs ...interface{}) {
			if len(prs) < 2 {
				cli.SendWrongParamCount()
			}

			cli.ListLPush(prs[0].(string), prs[1:])
		},
		true,
	}

	s.commands["RPUSH"] = command{
		true,
		func(cli *Client, prs ...interface{}) {
			if len(prs) < 2 {
				cli.SendWrongParamCount()
			}
			cli.ListRPush(prs[0].(string), prs[1:])
		},
		true,
	}

	s.commands["LPOP"] = command{
		true,
		func(cli *Client, prs ...interface{}) {
			if len(prs) < 1 {
				cli.SendWrongParamCount()
			}
			cli.ListLPop(prs[0].(string))
		},
		false,
	}

	s.commands["RPOP"] = command{
		true,
		func(cli *Client, prs ...interface{}) {
			if len(prs) < 1 {
				cli.SendWrongParamCount()
			}
			cli.ListRPop(prs[0].(string))
		},
		false,
	}

	s.commands["LINDEX"] = command{
		true,
		func(cli *Client, prs ...interface{}) {
			if len(prs) < 2 {
				cli.SendWrongParamCount()
			}
			cli.ListIndex(prs[0].(string), prs[1].(int))
		},
		false,
	}

	s.commands["LREM"] = command{
		true,
		func(cli *Client, prs ...interface{}) {
			if len(prs) < 2 {
				cli.SendWrongParamCount()
			}

			cli.ListRemove(prs[0].(string), prs[1].(int))
		},
		true,
	}

	s.commands["LINSERT"] = command{
		true,
		func(cli *Client, prs ...interface{}) {
			if len(prs) < 3 {
				cli.SendWrongParamCount()
			}

			cli.ListInsertAfter(prs[0].(string), prs[1].(int), prs[2])
		},
		true,
	}

	s.commands["LINSERT"] = command{
		true,
		func(cli *Client, prs ...interface{}) {
			if len(prs) < 3 {
				cli.SendWrongParamCount()
			}

			cli.ListInsertAfter(prs[0].(string), prs[1].(int), prs[2])
		},
		true,
	}
}

func (s *server) parseCommand(input *bufio.Reader) (name string, prs []interface{}, err error) {
	buf := make([]byte, 64*1024)
	c, err := input.Read(buf)
	if err != nil {
		println(err)
		return
	}
	src, err := resp.ParseRespString(string(buf[:c]))
	if err != nil {
		return "", nil, err
	}
	commandArray := src[0].([]interface{})
	name = strings.ToUpper(commandArray[0].(string))
	if len(commandArray) > 1 {
		prs = commandArray[1:]
		return name, prs, nil
	}

	prs = []interface{}{}
	return name, prs, nil
}
