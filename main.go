package main

import (
	"io/ioutil"
	"log"

	yaml "gopkg.in/yaml.v2"

	"flag"

	"github.com/rpoletaev/test_jn/jnserver"
	"github.com/xlab/closer"
)

func main() {
	defer closer.Close()
	config := getConfig()

	if config.Port <= 0 {
		println("Wrong port number. Default port number is 2020")
		config.Port = 2020
	}

	srv := jnserver.CreateServerWithConfig(&config)
	closer.Bind(srv.Stop)
	closer.Checked(srv.Run, true)
	closer.Hold()
}

func getConfig() jnserver.ServerConfig {
	var configPath string
	flag.StringVar(&configPath, "c", "", "-c /path/to/config.yaml")
	flag.Parse()

	if configPath == "" {
		log.Println("loading default config")
		return getDefaultConfig()
	}

	cb, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Error read config: %s\n %v", configPath, err)
	}

	var config jnserver.ServerConfig
	err = yaml.Unmarshal(cb, &config)
	if err != nil {
		log.Fatalf("Error config format: %v ", err)
	}

	return config
}

func getDefaultConfig() jnserver.ServerConfig {
	return jnserver.ServerConfig{}
}
