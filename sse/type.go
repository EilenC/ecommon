package sse

import (
	"fmt"
	"sync"
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
	messageChan chan string //推送消息通道
	createTime  int64       //连接创建时的时间戳(秒级)
}

// Packet server 消息包
type Packet struct {
	Message   Message `json:"message"` //发送内容消息体
	Zone      string  //类似区域概念,每个连接可以在不同区域中
	ID        string  `json:"id"` //连接ID,用于标识连接
	Broadcast bool    //是否广播
}

// Message sse消息内容
type Message struct {
	Event string //server 监听事件名称,必填
	Data  string //发送内容
}

// Format format server message
func (m Message) Format() string {
	return fmt.Sprintf("event: %s\ndata: %s\n\n", m.Event, m.Data)
}
