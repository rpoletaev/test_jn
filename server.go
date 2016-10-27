package test_juno

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

type ServerConfig struct {
	Port        int    `yaml:"port"`
	Requirepass bool   `yaml:"requirepass"`
	Password    string `yaml:"password"`
}

type server struct {
	cfg     *ServerConfig
	bases   []*Base
	clients []*Client
}

func StartServer() {
	defaultCfg := &ServerConfig{
		Port:        2020,
		Requirepass: true,
		Password:    "roma",
	}

	StartServerWithConfig(defaultCfg)
}

func StartServerWithConfig(c *ServerConfig) {
	server := &server{
		cfg:     c,
		clients: make([]*Client, 1),
	}

	if err := server.initBases(); err != nil {
		log.Fatal(err)
	}

	l, err := net.Listen("tcp", fmt.Sprint(":", c.Port))
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	log.Println("Listen on ", c.Port, "...")

	for {
		con, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go func(c net.Conn) {
			input := bufio.NewScanner(c)
			cli := &Client{
				authorized: !server.cfg.Requirepass,
				basenumber: 0,
				con:        c,
			}

			defer func() {
				c.Close()
			}()

			if err := server.Authorise(cli, input); err != nil {
				log.Fatalln(err)
				return
			}

			server.clients = append(server.clients, cli)

			for {
				input.Scan()
				fmt.Fprintln(cli.con, input.Text())
			}
		}(con)
	}
}

func (s *server) initBases() error {
	s.bases = make([]*Base, 1)
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

func (s *server) Authorise(cli *Client, input *bufio.Scanner) error {
	passCounter := 0
	for !cli.authorized {
		if passCounter == 3 {
			return fmt.Errorf("To many authorisation attempts")
		}

		fmt.Fprintln(cli.con, "Enter password:")
		input.Scan()
		cli.authorized = s.cfg.Password == input.Text()
		passCounter++
	}
	fmt.Fprintln(cli.con, "OK!")
	return nil
}
