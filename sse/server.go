// Package sse Package server implements Server-Sent Events, as specified in RFC 6202.
package sse

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// NewHub 返回SSE总hub
// reply 为nil时,不记录推送消息,否则会记录 设计用于返回推送数据 方便记录日志
func NewHub(reply chan string) *Hub {
	h := &Hub{
		cons:      make(map[string]map[string]Link),
		broadcast: make(chan Packet),
		block:     sync.Mutex{},
	}
	if reply != nil {
		h.reply = reply
	} else {
		h.reply = make(chan string)
		select {
		case <-h.reply:
		default:
		}
	}
	//started broadcast
	go func() {
		h.StartBroadcast()
	}()
	return h
}

// StartBroadcast 广播
func (hub *Hub) StartBroadcast() {
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					// 在这里处理 panic，并记录日志或采取其他措施
					hub.reply <- fmt.Sprintf("Server-Sent Events StartBroadcast Panic %+v", r)
					return
				}
			}()
			select {
			case message := <-hub.broadcast:
				hub.broadcastMessage(message)
			default:
			}
		}()
	}
}

// broadcastMessage 将信息广播所有房间与连接
func (hub *Hub) broadcastMessage(pkg Packet) {
	for zone, cons := range hub.cons {
		hub.broadcastZoneMessage(zone, pkg.Message, cons)
	}
}

// broadcastZoneMessage 域内广播
// zone 不为nil时,广播此所有连接
func (hub *Hub) broadcastZoneMessage(zone string, message Message, zones map[string]Link) {
	for id, b := range zones {
		select {
		case b.messageChan <- message.Format():
			hub.broadcastReply(zone, id, message)
		default:
		}
	}
}

// broadcastReply 广播消息后给chan推送,方便记录.
func (hub *Hub) broadcastReply(zone, id string, message Message) {
	if hub.reply != nil {
		select {
		case hub.reply <- fmt.Sprintf("%s:%s send [%s->%s]", zone, id, message.Event, message.Data):
		default:
		}
	}
}

// UnRegisterBlock 注销连接
// zone string 区域名称.
// id string http 连接名称.
func (hub *Hub) UnRegisterBlock(zone, id string) {
	hub.block.Lock()
	defer hub.block.Unlock()
	if links, ok := hub.cons[zone]; ok {
		if _, exist := links[id]; exist {
			delete(links, id)
		}
	}
	return
}

// RegisterBlock 注册SSE连接.
// zone string 区域名称 默认 default.
// uuid func() string 生成连接ID的函数,默认使用时间戳.
func (hub *Hub) RegisterBlock(w http.ResponseWriter, r *http.Request, zone string, uuid func() string) {
	if zone == "" {
		zone = "default"
	}
	if uuid == nil {
		uuid = func() string {
			return fmt.Sprintf("%d", time.Now().UnixNano())
		}
	}
	id := uuid()
	flusher, _ := w.(http.Flusher)
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	newBlock := Link{messageChan: make(chan string), createTime: time.Now().Unix()}
	hub.block.Lock()
	if hub.cons[zone] == nil {
		hub.cons[zone] = make(map[string]Link)
	}
	hub.cons[zone][id] = newBlock
	hub.block.Unlock()
	defer func() {
		close(newBlock.messageChan)
		hub.UnRegisterBlock(zone, id)
	}()
	fmt.Printf("in room id:%s\n", id)
	go func() {
		_ = hub.SendMessage(Packet{
			Message: Message{
				Event: "ping",
				Data:  fmt.Sprintf("%s->%s Connection Successful!", zone, id),
			}, Zone: zone, ID: id,
		})
	}()
	for {
		select {
		case message := <-newBlock.messageChan:
			// push message to client
			_, _ = io.WriteString(w, message)
			flusher.Flush()
		case <-r.Context().Done():
			// when "es.close()" is called, this loop operation will be ended.
			return
		}
	}
}

// SendMessage 发送消息 包含广播
// pkg Packet 消息包
func (hub *Hub) SendMessage(pkg Packet) error {
	lr := len(pkg.Zone)
	ld := len(pkg.ID)
	//全域广播,没有指定区域
	if pkg.Broadcast && lr == 0 && ld == 0 {
		hub.broadcast <- pkg
	}
	var (
		cons map[string]Link
		ok   bool
	)
	if lr != 0 {
		hub.block.Lock()
		cons, ok = hub.cons[pkg.Zone]
		hub.block.Unlock()
		if !ok {
			return fmt.Errorf("zone not exist")
		}
		if len(cons) == 0 {
			return fmt.Errorf("no connections are available")
		}
	}
	//域内广播
	if lr != 0 && pkg.Broadcast && ld == 0 {
		hub.broadcastZoneMessage(pkg.Zone, pkg.Message, cons)
	}
	//指定了连接ID 直接发送
	if len(pkg.ID) != 0 {
		var b Link
		hub.block.Lock()
		b, ok = cons[pkg.ID]
		hub.block.Unlock()
		if ok {
			b.messageChan <- pkg.Message.Format()
			return nil
		}
		return fmt.Errorf("push message to %s chan fail", pkg.ID)
	}
	return nil
}
