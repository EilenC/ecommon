package main

import (
	"fmt"
	"github.com/EilenC/ecommon/sse"
	"net/http"
	"time"
)

func main() {
	client := sse.NewClient("http://localhost:8080/sse", http.MethodGet, 3*time.Second)

	// 自定义事件处理逻辑
	client.OnEvent("ping", func(event *sse.Message) {
		fmt.Printf("ID:%s 收到 %s 事件: %s\n", event.ID, event.Event, event.Data)
	})

	// 自定义事件处理逻辑
	client.OnEvent("customEvent", func(event *sse.Message) {
		fmt.Printf("ID:%s 收到 %s 事件: %s\n", event.ID, event.Event, event.Data)
	})

	// 连接建立时的处理逻辑
	client.OnConnection(func() {
		fmt.Println("连接已建立")
	})

	// 连接错误时的处理逻辑
	client.OnError(func(err error) {
		fmt.Println("连接错误:", err)
	})

	// 启动 SSE 客户端
	client.Start()
}
