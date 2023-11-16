package main

import (
	"github.com/EilenC/ecommon/sse"
	"log"
	"net/http"
	"time"
)

func connect(id int) {
	// 创建 SSE 客户端
	client := sse.NewClient("http://10.8.16.42:8080/sse", http.MethodGet, 3*time.Second)

	client.OnConnection(func() {
		log.Printf("ID:[%d] 连接已建立\n", id)
	})
	client.OnDisconnect(func(err string) {
		log.Printf("ID:[%d] 连接已断开\n", id)
	})

	// 订阅事件，并设置回调函数
	client.SubscribeEvent("ping", func(message *sse.Message) {
		log.Printf("ID:[%d] 收到 %s 事件 数据: %s\n", id, message.Event, message.Data)
	})

	// 订阅事件，并设置回调函数
	client.SubscribeEvent("customEvent", func(message *sse.Message) {
		log.Printf("ID:[%d] 收到 %s 事件 数据: %s\n", id, message.Event, message.Data)
	})
	//go func() {
	//	time.Sleep(5 * time.Second)
	//	client.Stop()
	//	log.Println("client stop")
	//}()
	client.Start()
}

func main() {
	connect(0)
}
