package main

import (
	"fmt"
	tcproxy "tcp_proxy/tcproxy"
)

var (
	client_cnt  int
	message_cnt int
)

func main() {
	proxy := tcproxy.New("localhost:1111", "localhost:9999")

	proxy.OnNewClient(func(c *tcproxy.Client) {
		client_cnt++
		print_stat()
	})

	proxy.OnNewMessage(func(c *tcproxy.Client, message string) {
		message_cnt++
		print_stat()
	})

	proxy.OnClientConnectionClosed(func(c *tcproxy.Client, err error) {
		client_cnt--
		print_stat()
	})

	proxy.Listen()
}

func print_stat() {
	fmt.Printf("%d\t%d\r", client_cnt, message_cnt)
}
