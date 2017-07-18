package jnserver

import (
	"fmt"
	"log"
	"net"

	"strings"

	"sync"

	"github.com/rpoletaev/respio"
	"github.com/xlab/closer"
)

const (
	NotAuthorized = "User not authorized"
)

var mu sync.Mutex

// ServerConfig config
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
	config   ServerConfig
	clients  map[*client]bool
	bases    []*Base
	commands map[string]command
}

//CreateServer Creates server with user config
func CreateServer(c ServerConfig) *server {
	return &server{
		config:   c,
		clients:  make(map[*client]bool),
		bases:    nil,
		commands: make(map[string]command),
	}
}

func (s *server) Run() error {
	defer closer.Close()
	if err := s.initBases(); err != nil {
		return err
	}

	s.loadCommands()

	l, err := net.Listen("tcp", fmt.Sprint(":", s.config.Port))
	if err != nil {
		log.Fatal(err)
	}

	defer l.Close()

	log.Println("Listen on ", s.config.Port, "...")

	for {
		con, err := l.Accept()
		if err != nil {
			log.Printf("%v", err)
		}

		go func(nc net.Conn) {
			defer func() {
				nc.Close()
			}()

			cli := &client{
				authorized: s.config.Password == "",
				base:       s.bases[0],
				srv:        s,
				con:        nc,
				reader:     respio.NewReader(nc),
				writer:     respio.NewWriter(nc),
			}

			log.Println("client connected")
			for {
				cmdName, prs, err := cli.reader.ReadCommand()
				if err != nil {
					cli.writer.SendError(err.Error())
					continue
				}

				cmd, exist := s.commands[strings.ToUpper(cmdName)]
				if !exist {
					cli.sendUnknownCommand(cmdName)
					continue
				}

				if cmd.authRequired && !cli.authorized {
					cli.sendNotAuthenticated()
					continue
				}

				cmd.execute(cli, prs...)

				go func() {
					if cmd.writeToAof {
						cli.base.writeToAof(cmdName, prs...)
					}
				}()
			}
		}(con)
	}
}

// Stop stop server
func (s *server) Stop() {
	log.Println("stopping server")
	for _, b := range s.bases {
		b.Stop()
	}
}

func (s *server) initBases() error {
	s.bases = make([]*Base, 0)
	s.newBase()
	return nil
}

func (s *server) newBase() (baseNum int) {
	mu.Lock()
	defer mu.Unlock()
	baseNum = len(s.bases)
	newBase := &Base{Number: baseNum}
	newBase.Run()
	s.bases = append(s.bases, newBase)
	return baseNum
}

func (s *server) loadCommands() {

	s.loadServerCommands()

	s.loadStringCommands()

	s.loadListCommands()
}

func (s *server) loadServerCommands() {
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

	s.commands["NEWBASE"] = command{
		true,
		func(cli *client, prs ...interface{}) {
			baseNum := s.newBase()
			cli.base = s.bases[baseNum]
			cli.writer.SendRESPInt(int64(baseNum))
			cli.writer.Flush()
		},
		true,
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
}

func (s *server) loadStringCommands() {
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
}

func (s *server) loadListCommands() {
	s.commands["LPUSH"] = command{
		true,
		listLPush,
		true,
	}

	s.commands["RPUSH"] = command{
		true,
		listRPush,
		true,
	}

	s.commands["LPOP"] = command{
		true,
		listLPop,
		false,
	}

	s.commands["RPOP"] = command{
		true,
		listRPop,
		false,
	}

	s.commands["LINDEX"] = command{
		true,
		listIndex,
		false,
	}

	s.commands["LREM"] = command{
		true,
		listRemove,
		true,
	}

	s.commands["LINSERT"] = command{
		true,
		listInsert,
		true,
	}

	s.commands["LINSAFT"] = command{
		true,
		listInsertAfter,
		true,
	}

	s.commands["LLEN"] = command{
		true,
		listLength,
		true,
	}
}
