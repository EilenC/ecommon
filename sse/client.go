package sse

import (
	"io"
	"log"
	"net/http"
	"time"
)

// SubscribeEvent Subscribe to callbacks for events
func (c *Client) SubscribeEvent(eventName string, callback EventCallback) {
	c.eventCallbacks[eventName] = callback
}

func NewClient(url, method string, reconnectDelay time.Duration) *Client {
	if reconnectDelay == 0 {
		reconnectDelay = 3 * time.Second //default
	}
	return &Client{
		url:            url,
		method:         method,
		eventCallbacks: map[string]EventCallback{},
		client:         &http.Client{},
		reconnectDelay: reconnectDelay,
		stopSignal:     make(chan struct{}),
		exitSignal:     make(chan struct{}),
	}
}

// OnConnection callback function upon successful connection registration
// (not guaranteed to be faster than the first response data)
func (c *Client) OnConnection(handler func()) {
	c.connectionHandler = handler
}

// OnDisconnect callback function when connection is disconnected
func (c *Client) OnDisconnect(handler func(err string)) {
	c.disconnectHandler = handler
}

// OnExit callback function when client stop
func (c *Client) OnExit(handler func()) {
	c.exitHandler = handler
}

// Stop client stop connect
func (c *Client) Stop() {
	go func() {
		c.exitSignal <- struct{}{}
	}()
}

// connect client connect to server
func (c *Client) connect() {
	ticker := time.Tick(c.reconnectDelay)
	req, err := http.NewRequest(c.method, c.url, nil)
	if err != nil {
		log.Printf("create server connect fail:%+v\n", err)
		return
	}
	for {
		resp, err := c.client.Do(req)
		if err != nil {
			log.Printf("connecting to SSE %s server:%+v\n", c.url, err)
			log.Printf("reconnecting in %v...\n", c.reconnectDelay)
			<-ticker
			continue
		}
		if resp.StatusCode != http.StatusOK {
			log.Printf("http status code error %d \n", resp.StatusCode)
			log.Printf("reconnecting in %v...\n", c.reconnectDelay)
			<-ticker
			continue
		}
		if c.connectionHandler != nil {
			go c.connectionHandler()
		}
		go c.listenEvents(resp.Body)
		<-c.stopSignal

		if c.disconnectHandler != nil {
			go c.disconnectHandler("client stopped")
		}
		log.Printf("reconnecting in %v...\n", c.reconnectDelay)
		<-ticker
	}
}

// listenEvents listening for events sent by SSE servers
func (c *Client) listenEvents(body io.ReadCloser) {
	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(body)
	decoder := NewDecoder(body)
	for {
		message, err := decoder.Decode()
		if err != nil {
			// processing read error, possibly due to disconnected connection
			log.Println("reading from sse server:", err)
			c.stopSignal <- struct{}{}
			return
		}

		// call the callback function of the subscription
		if callback, ok := c.eventCallbacks[message.Event]; ok {
			go callback(message)
		}
	}
}

func (c *Client) Start() {
	go c.connect()
	// waiting for interrupt signal
	<-c.exitSignal
	if c.exitHandler != nil {
		c.exitHandler()
	}
}
