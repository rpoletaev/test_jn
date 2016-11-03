package test_jn

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"

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
	command      handler
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

			input := bufio.NewScanner(nc)
			cli := &Client{
				authorized: !server.Requirepass,
				base:       server.bases[0],
				srv:        server,
				con:        nc,
			}

			server.clients[cli] = !server.Requirepass
			log.Println("client connected")
			for {
				cmdName, prs := server.parseCommand(input)
				log.Println(cmdName, " ", prs)
				cmd, exist := server.commands[cmdName]
				if !exist {
					cli.SendUnknownCommand(cmdName)
				} else {
					if authenticated, ok := server.clients[cli]; ok && !authenticated && cmd.authRequired {
						cli.SendNotAuthenticated()
					} else {
						cmd.command(cli, prs...)
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
		func(cli *Client, prs ...string) {
			cmds := make([]string, len(s.commands))
			var counter int = 0
			for kc, _ := range s.commands {
				cmds[counter] = kc
				counter++
			}

			cli.reply(format_bulk_string(cmds...))
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
}

func (s *server) parseCommand(input *bufio.Scanner) (name string, prs []string) {
	input.Scan()
	fields := strings.Fields(input.Text())
	name = strings.ToUpper(fields[0])
	if len(fields) > 1 {
		prs = fields[1:len(fields)]
		return name, prs
	}

	prs = []string{}
	return name, prs
}
