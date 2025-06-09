package client

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	DefaultHost                    = "aprs.glidernet.org"
	PortFullFeed                   = 10152
	PortFiltered                   = 14580
	DefaultKeepAlive time.Duration = 240 * time.Second
)

// Client represents a connection to the OGN APRS server.
type Client struct {
	Username  string
	Filter    string
	Host      string
	KeepAlive time.Duration

	conn net.Conn
}

// New creates a new Client with the given callsign and optional filter.
func New(username string, filter string) *Client {
	return &Client{
		Username:  username,
		Filter:    filter,
		Host:      DefaultHost,
		KeepAlive: DefaultKeepAlive,
	}
}

// Connect establishes the tcp connection and performs the login.
func (c *Client) Connect() error {
	port := PortFullFeed
	if c.Filter != "" {
		port = PortFiltered
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", c.Host, port))
	if err != nil {
		return err
	}
	c.conn = conn

	login := fmt.Sprintf("user %s pass -1 vers go-ogn-client 0.0", c.Username)
	if c.Filter != "" {
		login += " filter " + c.Filter
	}
	login += "\n"
	_, err = c.conn.Write([]byte(login))
	return err
}

// Run reads packets from the server and calls the provided callback for each line.
// If autoreconnect is true, it will try to reconnect on network errors.
func (c *Client) Run(callback func(string), autoreconnect bool) error {
	for {
		if c.conn == nil {
			if err := c.Connect(); err != nil {
				if !autoreconnect {
					return err
				}
				time.Sleep(5 * time.Second)
				continue
			}
		}

		if err := c.loop(callback); err != nil {
			c.conn.Close()
			c.conn = nil
			if !autoreconnect {
				return err
			}
			time.Sleep(5 * time.Second)
		} else {
			return nil
		}
	}
}

func (c *Client) loop(callback func(string)) error {
	reader := bufio.NewReader(c.conn)
	keepAliveTicker := time.NewTicker(c.KeepAlive)
	defer keepAliveTicker.Stop()

	for {
		c.conn.SetReadDeadline(time.Now().Add(c.KeepAlive + 30*time.Second))
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		callback(line)

		select {
		case <-keepAliveTicker.C:
			c.conn.Write([]byte("#keepalive\n"))
		default:
		}
	}
}

// Disconnect closes the underlying connection.
func (c *Client) Disconnect() error {
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}
