package email

import (
	"fmt"
	"io"
	"mime"
	"strings"

	"github.com/EilenC/ecommon"
	"github.com/EilenC/ecommon/slices"
	"gopkg.in/gomail.v2"
)

type Mail struct {
	host     string //smtp server host
	port     int    //smtp server port
	username string //smtp auth username
	password string //smtp auth passworld

	sender string
	sem    chan struct{} // semaphore to limit concurrency

	callBack func(ids string, sendErr, linkErr error) //async send callback

	dialerFactory func() mailDialer
}

type mailDialer interface {
	Dial() (gomail.SendCloser, error)
	DialAndSend(...*gomail.Message) error
}

func (m *Mail) SetCallBack(callBack func(ids string, sendErr, linkErr error)) {
	m.callBack = callBack
}

func (m *Mail) SetSender(sender string) {
	m.sender = sender
}

func NewMail(host string, port int, userName, password, sender string) *Mail {
	m := &Mail{
		host:     host,
		port:     port,
		username: userName,
		password: password,
		sender:   sender,
		sem:      make(chan struct{}, 20), // Limit to 20 concurrent sends
	}
	m.dialerFactory = func() mailDialer {
		return gomail.NewDialer(m.host, m.port, m.username, m.password)
	}
	return m
}

// prepare Preparing to send email messages
func (m *Mail) prepare(emails []string, title, htmlBody string, attachment []string) (*gomail.Message, error) {
	message := gomail.NewMessage(gomail.SetCharset("UTF-8"), gomail.SetEncoding(gomail.Base64))
	message.SetAddressHeader("From", m.username, m.sender)
	//设置相关的参数
	subject := title
	message.SetHeader("To", emails...)
	message.SetHeader("Subject", subject)                                  //邮件标题
	message.SetBody("text/html", htmlBody)                                 //邮件正文内容
	for _, path := range slices.RemoveStringDuplicateUseCopy(attachment) { //去除重复的附件
		fileName := ecommon.GetAttachmentName(path, "|")
		bytes, getErr := ecommon.GetFileContent(path)
		if getErr != nil {
			return nil, fmt.Errorf("SendMail GetFileContent Param:%+v Err:%+v Continue File:[%s]", emails, getErr, path)
		}
		message.Attach(mime.QEncoding.Encode("UTF-8", fileName), gomail.SetCopyFunc(func(writer io.Writer) error {
			_, err := writer.Write(bytes)
			return err
		}))
	}
	return message, nil
}

func (m *Mail) callback(ids string, sendErr, linkErr error) {
	if m.callBack != nil {
		m.callBack(ids, sendErr, linkErr)
	}
}

// runSend 执行发送
func (m *Mail) send(dd mailDialer, ids string, message *gomail.Message) (sendErr, linkErr error) {
	d, err := dd.Dial()
	if err != nil {
		linkErr = fmt.Errorf("dials and authenticates to an SMTP server fail:[%+v] 发送邮件ID:%+v", err, ids)
		m.callback(ids, nil, linkErr)
		return nil, linkErr
	}
	sendErr = gomail.Send(d, message)
	if closerErr := d.Close(); closerErr != nil {
		linkErr = fmt.Errorf("sends the QUIT command and closes the connection to the server:[%+v]", closerErr)
	}
	m.callback(ids, sendErr, linkErr)
	return sendErr, linkErr
}

// AsyncSendMail 发送邮件
func (m *Mail) AsyncSendMail(emails, ids []string, title, htmlBody string, attachment []string) error {
	message, err := m.prepare(emails, title, htmlBody, attachment)
	if err != nil {
		return err
	}
	//发送邮件
	m.sem <- struct{}{} // Acquire semaphore
	go func() {
		defer func() { <-m.sem }() // Release semaphore
		m.send(m.dialerFactory(), strings.Join(ids, ","), message)
	}()
	return nil
}

// SendEmail 阻塞发送邮件
func (m *Mail) SendEmail(email, title, htmlBody string, attachment []string) error {
	message, err := m.prepare([]string{email}, title, htmlBody, attachment)
	if err != nil {
		return err
	}
	//发送邮件
	if err := m.dialerFactory().DialAndSend(message); err != nil {
		return err
	}
	return nil
}
