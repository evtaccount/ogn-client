package client

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	DefaultHost                    = "aprs.glidernet.org"
	PortFullFeed                   = 10152
	PortFiltered                   = 14580
	DefaultKeepAlive time.Duration = 240 * time.Second
	AppName                        = "go-ogn-client"
	AppVersion                     = "0.2"
)

// Client represents a connection to the OGN APRS server.
type Client struct {
	Username  string
	Filter    string
	Host      string
	KeepAlive time.Duration
	Logger    *log.Logger

	conn   net.Conn
	mu     sync.Mutex
	killed bool
}

// New creates a new Client with the given callsign and optional APRS filter.
// Filter examples:
//
//	"r/48.0/11.0/100"       — 100 km radius around lat=48.0, lon=11.0
//	"a/50/5/45/15"          — rectangular area (latN/lonW/latS/lonE)
//	"p/OGN/FLR"             — prefix filter for callsigns
//	"t/po"                  — type filter (position + object)
func New(username string, filter string) *Client {
	return &Client{
		Username:  username,
		Filter:    filter,
		Host:      DefaultHost,
		KeepAlive: DefaultKeepAlive,
	}
}

func (c *Client) log(format string, args ...interface{}) {
	if c.Logger != nil {
		c.Logger.Printf(format, args...)
	}
}

// Connect establishes a TCP connection and performs the APRS login.
func (c *Client) Connect() error {
	port := PortFullFeed
	if c.Filter != "" {
		port = PortFiltered
	}

	addr := fmt.Sprintf("%s:%d", c.Host, port)
	c.log("Connecting to %s", addr)

	conn, err := net.DialTimeout("tcp", addr, 30*time.Second)
	if err != nil {
		return fmt.Errorf("dial %s: %w", addr, err)
	}

	// Enable TCP keepalive at the OS level.
	if tc, ok := conn.(*net.TCPConn); ok {
		tc.SetKeepAlive(true)
		tc.SetKeepAlivePeriod(60 * time.Second)
	}

	c.mu.Lock()
	c.conn = conn
	c.killed = false
	c.mu.Unlock()

	login := fmt.Sprintf("user %s pass -1 vers %s %s", c.Username, AppName, AppVersion)
	if c.Filter != "" {
		login += " filter " + c.Filter
	}
	login += "\n"

	if _, err := conn.Write([]byte(login)); err != nil {
		conn.Close()
		return fmt.Errorf("send login: %w", err)
	}

	c.log("Logged in as %s (filter: %q)", c.Username, c.Filter)
	return nil
}

// Run reads packets from the server and calls callback for each line.
// If autoreconnect is true, it retries on network errors.
// TimedCallback, if non-nil, is called every KeepAlive interval.
func (c *Client) Run(callback func(string), autoreconnect bool) error {
	return c.RunWithTimedCallback(callback, nil, autoreconnect)
}

// RunWithTimedCallback is like Run but also invokes timedCallback on every
// keep-alive interval (useful for periodic stats, flushing, etc.).
func (c *Client) RunWithTimedCallback(callback func(string), timedCallback func(*Client), autoreconnect bool) error {
	for {
		c.mu.Lock()
		killed := c.killed
		c.mu.Unlock()
		if killed {
			return nil
		}

		if c.conn == nil {
			if err := c.Connect(); err != nil {
				c.log("Connect error: %v", err)
				if !autoreconnect {
					return err
				}
				time.Sleep(5 * time.Second)
				continue
			}
		}

		err := c.loop(callback, timedCallback)
		c.closeConn()
		if err != nil {
			c.log("Loop error: %v", err)
			if !autoreconnect {
				return err
			}
			time.Sleep(5 * time.Second)
		} else {
			return nil
		}
	}
}

func (c *Client) loop(callback func(string), timedCallback func(*Client)) error {
	reader := bufio.NewReader(c.conn)

	// Send keep-alive and read packets concurrently so that keep-alive
	// is not blocked by a pending read.
	errCh := make(chan error, 1)
	lineCh := make(chan string, 64)

	// Reader goroutine.
	go func() {
		for {
			c.conn.SetReadDeadline(time.Now().Add(c.KeepAlive + 30*time.Second))
			line, err := reader.ReadString('\n')
			if err != nil {
				errCh <- err
				return
			}
			line = strings.TrimSpace(line)
			if line != "" {
				lineCh <- line
			}
		}
	}()

	keepAliveTicker := time.NewTicker(c.KeepAlive)
	defer keepAliveTicker.Stop()

	for {
		select {
		case line := <-lineCh:
			callback(line)
		case <-keepAliveTicker.C:
			if _, err := c.conn.Write([]byte("#keepalive\n")); err != nil {
				return fmt.Errorf("send keepalive: %w", err)
			}
			c.log("Sent keepalive")
			if timedCallback != nil {
				timedCallback(c)
			}
		case err := <-errCh:
			return err
		}
	}
}

func (c *Client) closeConn() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}

// Disconnect gracefully closes the connection and stops Run.
func (c *Client) Disconnect() error {
	c.mu.Lock()
	c.killed = true
	conn := c.conn
	c.conn = nil
	c.mu.Unlock()

	if conn != nil {
		return conn.Close()
	}
	return nil
}
