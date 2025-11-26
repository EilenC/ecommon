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
}

func (m *Mail) SetCallBack(callBack func(ids string, sendErr, linkErr error)) {
	m.callBack = callBack
}

func (m *Mail) SetSender(sender string) {
	m.sender = sender
}

func NewMail(host string, port int, userName, password, sender string) *Mail {
	return &Mail{
		host:     host,
		port:     port,
		username: userName,
		password: password,
		sender:   sender,
		sem:      make(chan struct{}, 20), // Limit to 20 concurrent sends
	}
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

// runSend 执行发送
func (m *Mail) send(dd *gomail.Dialer, ids string, message *gomail.Message, errChan chan error) {
	var (
		err, sendErr error
		d            gomail.SendCloser
	)
	d, err = dd.Dial()
	if err != nil {
		errChan <- fmt.Errorf("dials and authenticates to an SMTP server fail:[%+v] 发送邮件ID:%+v", err, ids)
		return
	}
	defer func(d gomail.SendCloser, ids string, ddErr, sendErr error) {
		//<-sendEmailCount
		if ddErr == nil { //第三方链接正常，关闭链接
			closerErr := d.Close()
			if closerErr != nil {
				errChan <- fmt.Errorf("sends the QUIT command and closes the connection to the server:[%+v]", closerErr)
			}
		}
		m.callBack(ids, sendErr, ddErr)
		//if sendErr != nil || ddErr != nil {
		//若链接或发送失败,处理数据库中的数据
		//m.CallBack(eIDs, sendErr, ddErr)
		//if ddErr != nil {
		//	if strings.Contains(ddErr.Error(), "550 Mailbox not found") {
		//		//特殊情况接收邮箱不存在gomail: could not send email 1: 550 Mailbox not found or access denied 也返回发送成功
		//		e.Log.Infof("发送邮件失败 sysError 接收邮箱不存在或者未激活特殊情况 Err:[%s] EID:[%+v]", ddErr, eIDs)
		//	}
		//}
		//if sendErr != nil {
		//	e.Log.Infof("发送邮件失败 sysError Err:[%s] EID:[%+v]", sendErr, eIDs)
		//}
		//return
		//}
	}(d, ids, err, sendErr)
	sendErr = gomail.Send(d, message)
}

// AsyncSendMail 发送邮件
func (m *Mail) AsyncSendMail(emails, ids []string, title, htmlBody string, attachment []string) error {
	message, err := m.prepare(emails, title, htmlBody, attachment)
	if err != nil {
		return err
	}
	//发送邮件
	dd := gomail.NewDialer(m.host, m.port, m.username, m.password)
	m.sem <- struct{}{} // Acquire semaphore
	go func() {
		defer func() { <-m.sem }() // Release semaphore
		m.send(dd, strings.Join(ids, ","), message, nil)
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
	dd := gomail.NewDialer(m.host, m.port, m.username, m.password)
	if err := dd.DialAndSend(message); err != nil {
		return err
	}
	return nil
}
