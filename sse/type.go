package sse

import (
	"sync"
	"time"
)

const (
	EOF = "\n"
)

// Hub Global SSE Hub
// reply is nil, no record push message, otherwise it will record
type Hub struct {
	cons      map[string]map[string]Link
	broadcast chan Packet //all broadcast
	block     sync.Mutex  //block cons
	reply     chan string
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
	ID        string   `json:"id"` //连接ID,用于标识连接
	Broadcast bool     //是否广播
}

// Message server 消息内容
type Message struct {
	timestamp time.Time
	ID        string //消息ID,可选
	Event     string //server 监听事件名称,必填
	Data      string //发送内容
	Retry     string //重试
	Comment   string //注释
}
