package jnserver

import (
	"fmt"
	"log"
	"net"

	"strings"

	"strconv"

	"github.com/rpoletaev/respio"
	"github.com/xlab/closer"
)

const (
	NotAuthorized = "User not authorized"
)

type ServerConfig struct {
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
}

type command struct {
	authRequired bool
	execute      handler
	writeToAof   bool
}

type server struct {
	*ServerConfig
	clients  map[*client]bool
	bases    []*Base
	commands map[string]command
}

//CreateServer Creates server with default config
func CreateServer() *server {
	defaultCfg := &ServerConfig{
		Port: 2020,
	}

	return CreateServerWithConfig(defaultCfg)
}

//CreateServerWithConfig Creates server with user config
func CreateServerWithConfig(c *ServerConfig) *server {
	return &server{
		c,
		make(map[*client]bool),
		nil,
		make(map[string]command),
	}
}

func (server *server) Run() error {
	defer closer.Close()
	if err := server.initBases(); err != nil {
		// log.Fatal(err)
		return err
	}

	server.loadCommands()

	l, err := net.Listen("tcp", fmt.Sprint(":", server.Port))
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer l.Close()

	log.Println("Listen on ", server.Port, "...")

	for {
		con, err := l.Accept()
		if err != nil {
			// log.Fatal(err)
			return err
		}

		go func(nc net.Conn) {
			defer func() {
				nc.Close()
				log.Println("Exit from client routine")
			}()

			cli := &client{
				authorized: server.Password == "",
				base:       server.bases[0],
				srv:        server,
				con:        nc,
				reader:     respio.NewReader(nc),
				writer:     respio.NewWriter(nc),
			}

			server.clients[cli] = cli.authorized
			log.Println("client connected")
			for {
				cmdName, prs, err := cli.reader.ReadCommand()
				if err != nil {
					cli.writer.SendError(err.Error())
					return
				}
				cmd, exist := server.commands[strings.ToUpper(cmdName)]
				if !exist {
					println("unknown command")
					cli.sendUnknownCommand(cmdName)
				} else {
					if authenticated, ok := server.clients[cli]; ok && !authenticated && cmd.authRequired {
						cli.sendNotAuthenticated()
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

	s.commands["SELECT"] = command{
		true,
		selectDBCommand,
		false,
	}

	s.commands["CMDS"] = command{
		false,
		func(cli *client, prs ...interface{}) {
			cli.writer.SendArray(int64(len(s.commands)))
			for kc := range s.commands {
				cli.writer.SendBulkString(kc)
			}
			cli.writer.Flush()
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

	s.commands["EXPIRE"] = command{
		true,
		expireCommand,
		true,
	}

	s.commands["TTL"] = command{
		true,
		getTTLCommand,
		false,
	}
	//LIST COMMANDS
	s.commands["LPUSH"] = command{
		true,
		func(cli *client, prs ...interface{}) {
			if len(prs) < 2 {
				cli.sendWrongParamCount()
				return
			}

			key, err := getStringFromParam(prs[0])
			if err != nil {
				cli.sendError(err.Error())
				return
			}
			cli.ListLPush(key, prs[1:])
		},
		true,
	}

	s.commands["RPUSH"] = command{
		true,
		func(cli *client, prs ...interface{}) {
			if len(prs) < 2 {
				cli.sendWrongParamCount()
				return
			}

			key, err := getStringFromParam(prs[0])
			if err != nil {
				cli.sendError(err.Error())
				return
			}
			cli.ListRPush(key, prs[1:])
		},
		true,
	}

	s.commands["LPOP"] = command{
		true,
		func(cli *client, prs ...interface{}) {
			if len(prs) < 1 {
				cli.sendWrongParamCount()
				return
			}

			key, err := getStringFromParam(prs[0])
			if err != nil {
				cli.sendError(err.Error())
				return
			}
			cli.ListLPop(key)
		},
		false,
	}

	s.commands["RPOP"] = command{
		true,
		func(cli *client, prs ...interface{}) {
			if len(prs) < 1 {
				cli.sendWrongParamCount()
				return
			}

			key, err := getStringFromParam(prs[0])
			if err != nil {
				cli.sendError(err.Error())
				return
			}
			cli.ListRPop(key)
		},
		false,
	}

	s.commands["LINDEX"] = command{
		true,
		func(cli *client, prs ...interface{}) {
			if len(prs) < 2 {
				cli.sendWrongParamCount()
				return
			}

			key, err := getStringFromParam(prs[0])
			if err != nil {
				cli.sendError(err.Error())
				return
			}

			idx, err := strconv.Atoi(string(prs[1].([]byte)))
			if err != nil {
				cli.sendWrongParamType("Integer")
				return
			}
			cli.ListIndex(key, idx)
		},
		false,
	}

	s.commands["LREM"] = command{
		true,
		func(cli *client, prs ...interface{}) {
			if len(prs) < 2 {
				cli.sendWrongParamCount()
				return
			}

			key, err := getStringFromParam(prs[0])
			if err != nil {
				cli.sendError(err.Error())
				return
			}

			idx, err := strconv.Atoi(string(prs[1].([]byte)))
			if err != nil {
				cli.sendWrongParamType("Integer")
				return
			}

			cli.ListRemove(key, idx)
		},
		true,
	}

	s.commands["LINSERT"] = command{
		true,
		func(cli *client, prs ...interface{}) {
			if len(prs) < 3 {
				cli.sendWrongParamCount()
				return
			}

			key, err := getStringFromParam(prs[0])
			if err != nil {
				cli.sendError(err.Error())
				return
			}

			idx, err := strconv.Atoi(string(prs[1].([]byte)))
			if err != nil {
				cli.sendWrongParamType("Integer")
				return
			}

			cli.ListSetIndex(key, idx, prs[2])
		},
		true,
	}

	s.commands["LINSAFT"] = command{
		true,
		func(cli *client, prs ...interface{}) {
			if len(prs) < 3 {
				cli.sendWrongParamCount()
				return
			}

			key, err := getStringFromParam(prs[0])
			if err != nil {
				cli.sendError(err.Error())
				return
			}

			idx, err := strconv.Atoi(string(prs[1].([]byte)))
			if err != nil {
				cli.sendWrongParamType("Integer")
				return
			}
			cli.ListInsertAfter(key, idx, prs[2])
		},
		true,
	}

	s.commands["LLEN"] = command{
		true,
		func(cli *client, prs ...interface{}) {
			if len(prs) != 1 {
				cli.sendWrongParamCount()
				return
			}

			key, err := getStringFromParam(prs[0])
			if err != nil {
				cli.sendError(err.Error())
				return
			}

			cli.ListLength(key)
		},
		true,
	}
}
