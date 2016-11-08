package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/rpoletaev/test_jn/resp"
)

func main() {
	con, err := net.Dial("tcp", "localhost:2020")
	if err != nil {
		println(err.Error())
		return
	}
	defer con.Close()
	buf := make([]byte, 64*1024)

	for {
		fmt.Print(">")
		reader := bufio.NewReader(os.Stdin)
		cmd, _, _ := reader.ReadLine()
		fields := strings.Fields(string(cmd))
		respCmd := resp.FormatBulkStringArray(fields)
		con.Write(respCmd)
		con.Read(buf)
		response, _ := resp.ParseRespString(string(buf))
		printArray(response)
	}
}

func printArray(arr []interface{}) {
	for _, val := range arr {
		switch val.(type) {
		case string:
			println(val.(string))
			break
		case []interface{}:
			printArray(val.([]interface{}))
			break
		case int64:
			println("(integer) ", val.(string))
			break
		case error:
			println(val.(string))
			break
		default:
			break
		}
	}
}
