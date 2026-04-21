package email

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"gopkg.in/gomail.v2"
)

func TestNewMail(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		port     int
		userName string
		password string
		sender   string
	}{
		{
			name:     "Create mail client with valid params",
			host:     "smtp.example.com",
			port:     587,
			userName: "user@example.com",
			password: "password",
			sender:   "Sender Name",
		},
		{
			name:     "Create mail client with different port",
			host:     "smtp.gmail.com",
			port:     465,
			userName: "test@gmail.com",
			password: "secret",
			sender:   "Test Sender",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMail(tt.host, tt.port, tt.userName, tt.password, tt.sender)
			if got == nil {
				t.Error("NewMail() returned nil")
				return
			}
			if got.host != tt.host {
				t.Errorf("NewMail() host = %v, want %v", got.host, tt.host)
			}
			if got.port != tt.port {
				t.Errorf("NewMail() port = %v, want %v", got.port, tt.port)
			}
			if got.username != tt.userName {
				t.Errorf("NewMail() username = %v, want %v", got.username, tt.userName)
			}
			if got.password != tt.password {
				t.Errorf("NewMail() password = %v, want %v", got.password, tt.password)
			}
			if got.sender != tt.sender {
				t.Errorf("NewMail() sender = %v, want %v", got.sender, tt.sender)
			}
			if got.sem == nil {
				t.Error("NewMail() sem channel is nil")
			}
			if got.dialerFactory == nil || got.dialerFactory() == nil {
				t.Error("NewMail() dialerFactory is nil")
			}
		})
	}
}

func TestMail_SetCallBack(t *testing.T) {
	m := NewMail("smtp.test.com", 587, "user@test.com", "pass", "Sender")

	callbackCalled := false
	callback := func(ids string, sendErr, linkErr error) {
		callbackCalled = true
	}

	m.SetCallBack(callback)

	if m.callBack == nil {
		t.Error("SetCallBack() callback was not set")
	}

	// Test that callback can be called
	m.callBack("test-id", nil, nil)
	if !callbackCalled {
		t.Error("SetCallBack() callback was not called")
	}
}

func TestMail_SetSender(t *testing.T) {
	m := NewMail("smtp.test.com", 587, "user@test.com", "pass", "Original Sender")

	tests := []struct {
		name   string
		sender string
	}{
		{
			name:   "Set new sender",
			sender: "New Sender",
		},
		{
			name:   "Set empty sender",
			sender: "",
		},
		{
			name:   "Set sender with special chars",
			sender: "Sender 测试",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.SetSender(tt.sender)
			if m.sender != tt.sender {
				t.Errorf("SetSender() sender = %v, want %v", m.sender, tt.sender)
			}
		})
	}
}

func TestMail_prepare(t *testing.T) {
	m := NewMail("smtp.test.com", 587, "user@test.com", "pass", "Test Sender")
	tempDir := t.TempDir()

	// Create test attachment files
	attachment1 := filepath.Join(tempDir, "test1.txt")
	if err := os.WriteFile(attachment1, []byte("attachment 1 content"), 0644); err != nil {
		t.Fatalf("Failed to create test attachment: %v", err)
	}

	attachment2 := filepath.Join(tempDir, "test2.txt")
	if err := os.WriteFile(attachment2, []byte("attachment 2 content"), 0644); err != nil {
		t.Fatalf("Failed to create test attachment: %v", err)
	}

	tests := []struct {
		name       string
		emails     []string
		title      string
		htmlBody   string
		attachment []string
		wantErr    bool
	}{
		{
			name:       "Prepare email without attachments",
			emails:     []string{"test@example.com"},
			title:      "Test Email",
			htmlBody:   "<h1>Hello</h1>",
			attachment: []string{},
			wantErr:    false,
		},
		{
			name:       "Prepare email with one attachment",
			emails:     []string{"test@example.com"},
			title:      "Test Email with Attachment",
			htmlBody:   "<p>Email body</p>",
			attachment: []string{attachment1},
			wantErr:    false,
		},
		{
			name:       "Prepare email with multiple attachments",
			emails:     []string{"test1@example.com", "test2@example.com"},
			title:      "Multiple Recipients",
			htmlBody:   "<p>Body</p>",
			attachment: []string{attachment1, attachment2},
			wantErr:    false,
		},
		{
			name:       "Prepare email with duplicate attachments",
			emails:     []string{"test@example.com"},
			title:      "Duplicate Attachments",
			htmlBody:   "<p>Body</p>",
			attachment: []string{attachment1, attachment1, attachment2},
			wantErr:    false,
		},
		{
			name:       "Prepare email with non-existent attachment",
			emails:     []string{"test@example.com"},
			title:      "Invalid Attachment",
			htmlBody:   "<p>Body</p>",
			attachment: []string{"/non/existent/file.txt"},
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := m.prepare(tt.emails, tt.title, tt.htmlBody, tt.attachment)
			if (err != nil) != tt.wantErr {
				t.Errorf("prepare() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Error("prepare() returned nil message")
			}
			if !tt.wantErr && len(tt.attachment) > 0 {
				if _, err := got.WriteTo(io.Discard); err != nil {
					t.Errorf("prepare() message WriteTo error = %v", err)
				}
			}
		})
	}
}

func TestMail_send(t *testing.T) {
	message, err := NewMail("host", 25, "from@example.com", "", "sender").prepare(
		[]string{"to@example.com"},
		"title",
		"body",
		nil,
	)
	if err != nil {
		t.Fatalf("prepare() error = %v", err)
	}

	tests := []struct {
		name        string
		dialErr     error
		sendErr     error
		closeErr    error
		wantSendErr bool
		wantLinkErr bool
	}{
		{name: "success"},
		{name: "dial error", dialErr: errors.New("dial failed"), wantLinkErr: true},
		{name: "send error", sendErr: errors.New("send failed"), wantSendErr: true},
		{name: "close error", closeErr: errors.New("close failed"), wantLinkErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMail("host", 25, "from@example.com", "", "sender")
			callbackCalled := false
			m.SetCallBack(func(ids string, sendErr, linkErr error) {
				callbackCalled = true
				if ids != "1,2" {
					t.Errorf("callback ids = %v, want 1,2", ids)
				}
				if (sendErr != nil) != tt.wantSendErr {
					t.Errorf("callback sendErr = %v, wantSendErr %v", sendErr, tt.wantSendErr)
				}
				if (linkErr != nil) != tt.wantLinkErr {
					t.Errorf("callback linkErr = %v, wantLinkErr %v", linkErr, tt.wantLinkErr)
				}
			})

			sendErr, linkErr := m.send(&fakeDialer{
				dialErr: dialErrOrNil(tt.dialErr),
				closer:  &fakeSendCloser{sendErr: tt.sendErr, closeErr: tt.closeErr},
			}, "1,2", message)
			if (sendErr != nil) != tt.wantSendErr {
				t.Errorf("send() sendErr = %v, wantSendErr %v", sendErr, tt.wantSendErr)
			}
			if (linkErr != nil) != tt.wantLinkErr {
				t.Errorf("send() linkErr = %v, wantLinkErr %v", linkErr, tt.wantLinkErr)
			}
			if !callbackCalled {
				t.Error("send() callback was not called")
			}
		})
	}
}

func TestMail_sendWithoutCallback(t *testing.T) {
	m := NewMail("host", 25, "from@example.com", "", "sender")
	message, err := m.prepare([]string{"to@example.com"}, "title", "body", nil)
	if err != nil {
		t.Fatalf("prepare() error = %v", err)
	}
	if sendErr, linkErr := m.send(&fakeDialer{closer: &fakeSendCloser{}}, "id", message); sendErr != nil || linkErr != nil {
		t.Fatalf("send() errors = %v, %v", sendErr, linkErr)
	}
}

func TestMail_AsyncSendMail(t *testing.T) {
	m := NewMail("host", 25, "from@example.com", "", "sender")
	done := make(chan struct{})
	m.dialerFactory = func() mailDialer {
		return &fakeDialer{closer: &fakeSendCloser{}}
	}
	m.SetCallBack(func(ids string, sendErr, linkErr error) {
		if ids != "42" {
			t.Errorf("callback ids = %v, want 42", ids)
		}
		if sendErr != nil || linkErr != nil {
			t.Errorf("callback errors = %v, %v", sendErr, linkErr)
		}
		close(done)
	})

	if err := m.AsyncSendMail([]string{"to@example.com"}, []string{"42"}, "title", "body", nil); err != nil {
		t.Fatalf("AsyncSendMail() error = %v", err)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("AsyncSendMail() callback timed out")
	}
}

func TestMail_AsyncSendMailPrepareError(t *testing.T) {
	m := NewMail("host", 25, "from@example.com", "", "sender")
	if err := m.AsyncSendMail([]string{"to@example.com"}, []string{"1"}, "title", "body", []string{"/not/exist"}); err == nil {
		t.Fatal("AsyncSendMail() expected prepare error")
	}
}

func TestMail_SendEmail(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{name: "success"},
		{name: "dial and send error", err: errors.New("send failed"), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMail("host", 25, "from@example.com", "", "sender")
			fake := &fakeDialer{dialAndSendErr: tt.err}
			m.dialerFactory = func() mailDialer {
				return fake
			}
			err := m.SendEmail("to@example.com", "title", "body", nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && fake.dialAndSendCount != 1 {
				t.Errorf("DialAndSend count = %d, want 1", fake.dialAndSendCount)
			}
		})
	}
}

func TestMail_SendEmailPrepareError(t *testing.T) {
	m := NewMail("host", 25, "from@example.com", "", "sender")
	if err := m.SendEmail("to@example.com", "title", "body", []string{"/not/exist"}); err == nil {
		t.Fatal("SendEmail() expected prepare error")
	}
}

func dialErrOrNil(err error) error {
	return err
}

type fakeDialer struct {
	dialErr          error
	dialAndSendErr   error
	closer           *fakeSendCloser
	dialAndSendCount int
}

func (d *fakeDialer) Dial() (gomail.SendCloser, error) {
	if d.dialErr != nil {
		return nil, d.dialErr
	}
	if d.closer == nil {
		d.closer = &fakeSendCloser{}
	}
	return d.closer, nil
}

func (d *fakeDialer) DialAndSend(messages ...*gomail.Message) error {
	d.dialAndSendCount++
	if d.dialAndSendErr != nil {
		return d.dialAndSendErr
	}
	for _, message := range messages {
		if _, err := message.WriteTo(io.Discard); err != nil {
			return err
		}
	}
	return nil
}

type fakeSendCloser struct {
	mu       sync.Mutex
	sendErr  error
	closeErr error
	from     string
	to       []string
	body     strings.Builder
	closed   bool
}

func (s *fakeSendCloser) Send(from string, to []string, msg io.WriterTo) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.from = from
	s.to = append([]string(nil), to...)
	_, _ = msg.WriteTo(&s.body)
	return s.sendErr
}

func (s *fakeSendCloser) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	return s.closeErr
}
