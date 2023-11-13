package sse

import (
	"io"
	"log"
	"net/http"
	"time"
)

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
