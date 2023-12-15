package sse

import (
	"bufio"
	"net/http"
	"sync"
	"time"
)

const (
	EOF = "\n"
)

// Hub Global SSE Hub
// reply is nil, no record push message, otherwise it will record
type Hub struct {
	cons           map[string]map[string]Link
	broadcast      chan Packet //all broadcast
	block          sync.Mutex  //block cons
	log            Log
	ConnectedFunc  func(clientID string) //连接建立时的处理逻辑
	DisconnectFunc func(clientID string) //连接建立时的处理逻辑
}

// Link server 连接
// messageChan 推送消息通道
// createTime 连接创建时的时间戳(秒级)
type Link struct {
	messageChan chan *Message //推送消息通道
	createTime  int64         //连接创建时的时间戳(秒级)
}

// Packet server 消息包
type Packet struct {
	Message   *Message `json:"message"` //发送内容消息体
	Zone      string   //类似区域概念,每个连接可以在不同区域中
	ClientID  string   `json:"client_id"` //连接ID,用于标识连接
	Broadcast bool     //是否广播
}

// Message 消息内容
type Message struct {
	timestamp time.Time
	ID        string //消息ID,可选
	Event     string //server 监听事件名称,必填
	Data      string //发送内容
	Retry     string //重试
	Comment   string //注释
}

// Decoder sse 解码器
type Decoder struct {
	reader *bufio.Reader
}

// EventCallback Define the type of SSE event subscription callback function
type EventCallback func(message *Message)

type Client struct {
	url               string
	method            string
	eventCallbacks    map[string]EventCallback
	client            *http.Client
	reconnectDelay    time.Duration
	connectionHandler func()
	disconnectHandler func(err string)
	exitHandler       func()
	stopSignal        chan struct{}
	exitSignal        chan struct{}
}
