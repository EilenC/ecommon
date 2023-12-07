package main

import (
	"fmt"
	"github.com/EilenC/ecommon/email"
)

const (
	userName  = ""
	passworld = ""
	toMail    = ""
)

func asyncSend(m *email.Mail, email string) {
	err := m.AsyncSendMail([]string{email}, []string{"1"}, "async test email", "this is a test email",
		[]string{"https://baike.seekhill.com/uploads/202106/1624354355xY7cLkuE_s.png"})
	m.SetCallBack(func(ids string, sendErr, linkErr error) {
		fmt.Println("send successful callback ids", ids)
	})
	if err != nil {
		panic(err)
	}
}

func send(m *email.Mail, email string) {
	err := m.SendEmail(email, "test email", "this is a test email", nil)
	if err != nil {
		panic(err)
	}
}

func main() {
	m := email.NewMail("smtp.exmail.qq.com", 465, userName, passworld, "")
	asyncSend(m, toMail)
	send(m, toMail)
	select {}
}
