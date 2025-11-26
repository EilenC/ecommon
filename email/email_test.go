package email

import (
	"os"
	"path/filepath"
	"testing"
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
		})
	}
}

// Note: We don't test AsyncSendMail and SendEmail because they require a real SMTP server
// These functions would need integration tests with a test SMTP server
// Similarly, HTTP attachment downloads are tested in the file_test.go tests
