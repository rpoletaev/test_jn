package main

import (
	"github.com/rpoletaev/test_jn"
	"github.com/xlab/closer"
)

func main() {
	defer closer.Close()
	srv := test_jn.CreateServer()
	closer.Bind(srv.Stop)
	srv.Run()
	// time.Sleep(5 * time.Second)
	closer.Hold()
}
