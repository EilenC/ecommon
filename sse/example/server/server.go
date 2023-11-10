package main

import (
	"encoding/json"
	"fmt"
	"github.com/EilenC/ecommon/sse"
	"net/http"
	"time"
)

func main() {
	index := func(w http.ResponseWriter, r *http.Request) {
		page := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Server-Sent Events</title>
</head>
<body>
<h1>Server-Sent Events</h1>
<div>
    <button onclick="connectSSE()">建立 SSE 连接</button>
    <button onclick="closeSSE()">断开 SSE 连接</button>
</div>
<div style="margin-top: 20px;">
    接收ID:<input type="text" id="user" name="user">
    发送内容:<input type="text" id="content" name="content">
    <button onclick="send()">发送内容至指定ID</button>
    <button onclick="broadcast()">广播消息</button>
</div>
<br />
<br />
<div id="message"></div>

<script>
    const messageElement = document.getElementById('message')

    let eventSource

    // 建立 SSE 连接
    const connectSSE = () => {
        eventSource = new EventSource('/sse')

        eventSource.addEventListener('ping', (event) => {
            messageElement.innerHTML += "ping:"+event.data+"<br />"
        })

        // 监听消息事件
        eventSource.addEventListener('customEvent', (event) => {
            messageElement.innerHTML += event.data + '<br />'
        })

        eventSource.onopen = () => {
            messageElement.innerHTML += "SSE 连接成功<br />"
        }

        eventSource.onerror = () => {
            messageElement.innerHTML += "SSE 连接错误<br />"
        }
    }

    // 断开 SSE 连接
    const closeSSE = () => {
        eventSource.close()
        messageElement.innerHTML += "SSE 连接关闭<br />"
    }

    const broadcast = () => {
        var content = document.getElementById("content").value;
        fetch('/broadcast?content='+content)
            .then(response => response.text()) // 或者使用 response.json() 来解析 JSON 响应
            .then(data => {
                console.log(data); // 处理响应数据
            })
            .catch(error => {
                console.error('发生错误：', error); // 处理错误
            });
    }

    const send = () => {
        var userID = document.getElementById("user").value;
        var content = document.getElementById("content").value;
        fetch('/send?user='+ userID+'&content='+content)
            .then(response => response.text()) // 或者使用 response.json() 来解析 JSON 响应
            .then(data => {
				data != '' ? alert(data):console.log(data); // 处理响应数据
            })
            .catch(error => {
                console.error('发生错误：', error); // 处理错误
            });
    }
</script>
</body>
</html>`
		_, _ = w.Write([]byte(page))
	}
	var h *sse.Hub
	h = sse.NewHub(nil)
	send := func(w http.ResponseWriter, r *http.Request) {
		user := r.FormValue("user")
		content := r.FormValue("content")
		msg := make(map[string]interface{})
		msg["msg"] = content
		msg["time"] = time.Now().UnixMilli()
		b, _ := json.Marshal(msg)
		err := h.SendMessage(sse.Packet{
			Message: &sse.Message{
				Event: "customEvent",
				Data:  string(b),
			},
			Zone:      "default",
			ClientID:  user,
			Broadcast: user == "",
		})
		if err != nil {
			fmt.Printf("send message fail %+v\n", err)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
	}

	loop := func(w http.ResponseWriter, r *http.Request) {
		for {
			msg := fmt.Sprintf("loop %+v\n", time.Now().Unix())
			fmt.Printf(msg)
			h.SendMessage(sse.Packet{
				Message: &sse.Message{
					Event: "customEvent",
					Data:  msg,
				},
				Zone:      "",
				ClientID:  "",
				Broadcast: true,
			})
			time.Sleep(time.Second * 1)
		}
	}
	http.HandleFunc("/", index)
	http.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
		h.RegisterBlock(w, r, "default", nil)
	})
	http.HandleFunc("/broadcast", send)
	http.HandleFunc("/send", send)
	http.HandleFunc("/loop", loop)
	port := "8080"
	fmt.Printf("Listening on :%s\n", port)
	_ = http.ListenAndServe(":"+port, nil)
}
