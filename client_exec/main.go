package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	con, err := net.Dial("tcp", "localhost:2020")
	if err != nil {
		println(err.Error())
		return
	}
	defer con.Close()

	for {
		fmt.Print(">")
		reader := bufio.NewReader(os.Stdin)
		txt, _, _ := reader.ReadLine()
		fmt.Fprintln(con, string(txt))

		line, _ := bufio.NewReader(con).ReadString('\n')
		fmt.Println(line)
	}
}
