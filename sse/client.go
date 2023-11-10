package sse

import (
	"bufio"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// Client server-sent events client
type Client struct {
	url               string
	method            string
	eventHandlers     map[string]func(event *Message)
	connectionHandler func()
	errorHandler      func(err error)
	client            *http.Client
	replyTime         time.Duration
	StopReplySignal   chan struct{}
}

// NewClient server-sent events client
func NewClient(url, method string, replyTime time.Duration) *Client {
	if replyTime == 0 {
		replyTime = 3 * time.Second //default
	}
	if method == "" {
		method = http.MethodGet
	}
	return &Client{
		url:             url,
		eventHandlers:   make(map[string]func(event *Message)),
		client:          &http.Client{},
		replyTime:       replyTime,
		StopReplySignal: make(chan struct{}),
	}
}

// OnEvent subscribe event
func (c *Client) OnEvent(eventType string, handler func(event *Message)) {
	c.eventHandlers[eventType] = handler
}

// OnConnection register connected callback
func (c *Client) OnConnection(handler func()) {
	c.connectionHandler = handler
}

// OnError register connected callback
func (c *Client) OnError(handler func(err error)) {
	c.errorHandler = handler
}

// Start client connect server
func (c *Client) Start() {
	c.connect()

	ticker := time.Tick(c.replyTime)
	for {
		select {
		case <-ticker:
			c.connect()
		case <-c.StopReplySignal:
			// Stop loop and exit
			return
		}
	}
}

// connect server
func (c *Client) connect() {
	req, err := http.NewRequest(http.MethodGet, c.url, nil)
	if err != nil {
		log.Printf("create server connect fail:%+v\n", err)
		return
	}
	resp, err := c.client.Do(req)
	if err != nil {
		log.Printf("connect server fail:%+v\n", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("server response status code:%v fail\n", resp.StatusCode)
		return
	}
	c.handleConnectStream(resp.Body)
	log.Println("server connection disconnected, attempting to reconnect...")
	return
}

// Decoder
type Decoder struct {
	reader *bufio.Reader
}

// NewDecoder
func NewDecoder(reader io.Reader) *Decoder {
	return &Decoder{
		reader: bufio.NewReader(reader),
	}
}

// Decode
func (d *Decoder) Decode() (*Message, error) {
	event := &Message{}

	for {
		line, err := d.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = strings.TrimSpace(line)

		if line == "" {
			return event, nil
		}

		if strings.HasPrefix(line, "event:") {
			event.Event = strings.TrimSpace(line[6:])
		} else if strings.HasPrefix(line, "data:") {
			event.Data += strings.TrimSpace(line[5:]) + "\n"
		} else if strings.HasPrefix(line, "id:") {
			event.ID = strings.TrimSpace(line[3:])
		} else if strings.HasPrefix(line, "retry:") {
			event.Retry = strings.TrimSpace(line[6:])
		} else if strings.HasPrefix(line, ":") {
			event.Comment = strings.TrimSpace(line[1:])
		}
	}
}

// handleSSEStream
func (c *Client) handleConnectStream(body io.ReadCloser) {
	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(body)

	decoder := NewDecoder(body)
	for {
		event, err := decoder.Decode()
		if err != nil {
			log.Println("SSE 解码失败:", err)
			break
		}

		if event.Event == "connection_established" {
			if c.connectionHandler != nil {
				c.connectionHandler()
			}
			continue
		}

		handler, ok := c.eventHandlers[event.Event]
		if ok {
			handler(event)
		}
	}
}
