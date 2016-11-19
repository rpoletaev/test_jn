package main

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"

	"github.com/rpoletaev/test_jn/jnserver"
	"github.com/xlab/closer"
)

func main() {
	defer closer.Close()
	config := jnserver.ServerConfig{}
	configBytes, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		println("Error loading config.yaml")
		srv := jnserver.CreateServer()
		closer.Bind(srv.Stop)
		closer.Checked(srv.Run, true)
		closer.Hold()
	} else {
		yamlError := yaml.Unmarshal(configBytes, &config)
		if yamlError != nil {
			panic(yamlError)
		}

		if config.Port <= 0 {
			println("Wrong port number. Default port number is 2020")
			config.Port = 2020
		}

		srv := jnserver.CreateServerWithConfig(&config)
		closer.Bind(srv.Stop)
		closer.Checked(srv.Run, true)
		closer.Hold()
	}

}
