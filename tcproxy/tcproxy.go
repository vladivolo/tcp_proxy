package tcp_proxy

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"
)

// Client holds info about connection
type Client struct {
	in_conn  net.Conn
	out_conn net.Conn
	Server   *server
}

// TCP server
type server struct {
	address                  string // Address to open connection: localhost:9999
	dial_address             string
	onNewClientCallback      func(c *Client)
	onClientConnectionClosed func(c *Client, err error)
	onNewMessage             func(c *Client, message string)
}

func (c *Client) connect() error {
	t := 1 * time.Second

	for i := 0; i < 10; i++ {
		time.Sleep(t)

		conn, err := net.Dial("tcp", c.Server.DialAddr())
		if err != nil {
			log.Println("connect: ", err)
			//t *= 2
			continue
		}

		if c.out_conn != nil {
			c.out_conn.Close()
		}
		c.out_conn = conn

		log.Println("Connected!!!")
		return nil
	}

	return fmt.Errorf("EConnAborted")
}

// Read client data from channel
func (c *Client) listen() {
	var err error

	defer c.in_conn.Close()
	defer c.Server.onClientConnectionClosed(c, err)

	if err := c.connect(); err != nil {
		c.Send(err.Error())
		return
	}

	log.Println("Connect success: ", c.Server.DialAddr())

	reader := bufio.NewReader(c.in_conn)
	for {
		// Read Client cmd
		message, err := reader.ReadString('\n')
		if err != nil {
			c.out_conn.Close()
			c.Server.onClientConnectionClosed(c, err)
			return
		}
		c.Server.onNewMessage(c, message)

		// Write client cmd to server
	rewrite:
		log.Println("Try write to server: ", message)
		_, err = c.out_conn.Write([]byte(message))
		if err != nil {
			if err := c.connect(); err != nil {
				c.Send(err.Error())
				return
			}
			goto rewrite
		}

		// Read Server response
		log.Println("Try read from server: start")
		srv_reader := bufio.NewReader(c.out_conn)
		reply, err := srv_reader.ReadString('\n')
		if err != nil {
			if err := c.connect(); err != nil {
				c.Send(err.Error())
				return
			}
			goto rewrite
		}

		// Write response to Client
		err = c.Send(reply)
		if err != nil {
			c.out_conn.Close()
			return
		}
	}
}

// Send text message to client
func (c *Client) Send(message string) error {
	_, err := c.in_conn.Write([]byte(message))
	return err
}

// Send bytes to client
func (c *Client) SendBytes(b []byte) error {
	_, err := c.in_conn.Write(b)
	return err
}

func (c *Client) InConn() net.Conn {
	return c.in_conn
}

func (c *Client) OutConn() net.Conn {
	return c.out_conn
}

func (c *Client) InClose() error {
	return c.in_conn.Close()
}

func (c *Client) OutClose() error {
	return c.in_conn.Close()
}

func (s *server) DialAddr() string {
	return s.dial_address
}

// Called right after server starts listening new client
func (s *server) OnNewClient(callback func(c *Client)) {
	s.onNewClientCallback = callback
}

// Called right after connection closed
func (s *server) OnClientConnectionClosed(callback func(c *Client, err error)) {
	s.onClientConnectionClosed = callback
}

// Called when Proxy receives new message
func (s *server) OnNewMessage(callback func(c *Client, message string)) {
	s.onNewMessage = callback
}

// Start network server
func (s *server) Listen() {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		log.Fatal("Error starting TCP server.")
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Accept() return ", err)
			continue
		}
		client := &Client{
			in_conn: conn,
			Server:  s,
		}
		go client.listen()
		s.onNewClientCallback(client)
	}
}

// Creates new tcp server instance
func New(address string, dial string) *server {
	log.Println("Creating server with address", address)
	server := &server{
		address:      address,
		dial_address: dial,
	}

	server.OnNewClient(func(c *Client) {})
	server.OnNewMessage(func(c *Client, message string) {})
	server.OnClientConnectionClosed(func(c *Client, err error) {})

	return server
}
