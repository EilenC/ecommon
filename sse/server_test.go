package sse

import (
	"testing"
	"time"
)

func TestMessage_Format(t *testing.T) {
	type fields struct {
		timestamp time.Time
		ID        string
		Event     string
		Data      string
		Retry     string
		Comment   string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				timestamp: time.Time{},
				ID:        "123",
				Event:     "abc",
				Data:      "111",
				Retry:     "",
				Comment:   "",
			},
			want:    "id: 123\ndata: 111\nevent: abc\n\n",
			wantErr: false,
		},
		{
			name: "with retry",
			fields: fields{
				timestamp: time.Time{},
				ID:        "456",
				Event:     "test",
				Data:      "data content",
				Retry:     "3000",
				Comment:   "",
			},
			want:    "id: 456\ndata: data content\nevent: test\nretry: 3000\n\n",
			wantErr: false,
		},
		{
			name: "with comment only",
			fields: fields{
				timestamp: time.Time{},
				ID:        "",
				Event:     "",
				Data:      "",
				Retry:     "",
				Comment:   "heartbeat",
			},
			want:    ": heartbeat\n\n",
			wantErr: false,
		},
		{
			name: "empty data and comment",
			fields: fields{
				timestamp: time.Time{},
				ID:        "789",
				Event:     "empty",
				Data:      "",
				Retry:     "",
				Comment:   "",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				timestamp: tt.fields.timestamp,
				ID:        tt.fields.ID,
				Event:     tt.fields.Event,
				Data:      tt.fields.Data,
				Retry:     tt.fields.Retry,
				Comment:   tt.fields.Comment,
			}
			got, err := m.Format()
			if (err != nil) != tt.wantErr {
				t.Errorf("Format() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				gotMsg := got.String()
				if len(gotMsg) != len(tt.want) || gotMsg != tt.want {
					t.Errorf("Format() got = %s, want %s", gotMsg, tt.want)
				}
			}
		})
	}
}

func TestNewHub(t *testing.T) {
	tests := []struct {
		name string
		log  Log
	}{
		{
			name: "Create hub with nil log",
			log:  nil,
		},
		{
			name: "Create hub with mock log",
			log:  &mockLog{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewHub(tt.log)
			if got == nil {
				t.Error("NewHub() returned nil")
				return
			}
			if got.cons == nil {
				t.Error("NewHub() cons is nil")
			}
			if got.broadcast == nil {
				t.Error("NewHub() broadcast channel is nil")
			}
			if got.log != tt.log {
				t.Errorf("NewHub() log = %v, want %v", got.log, tt.log)
			}
		})
	}
}

func TestHub_getClientID(t *testing.T) {
	hub := NewHub(nil)

	tests := []struct {
		name   string
		length int
	}{
		{
			name:   "Generate 16 char ID",
			length: 16,
		},
		{
			name:   "Generate 8 char ID",
			length: 8,
		},
		{
			name:   "Generate 32 char ID",
			length: 32,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hub.getClientID(tt.length)
			if len(got) != tt.length {
				t.Errorf("getClientID() length = %v, want %v", len(got), tt.length)
			}

			// Test uniqueness
			got2 := hub.getClientID(tt.length)
			if got == got2 {
				t.Error("getClientID() generated duplicate ID")
			}
		})
	}
}

func TestHub_UnRegisterBlock(t *testing.T) {
	hub := NewHub(nil)

	// Setup: create a connection
	zone := "test-zone"
	id := "test-id"
	hub.cons[zone] = make(map[string]Link)
	hub.cons[zone][id] = Link{
		messageChan: make(chan *Message),
		allowPush:   make(chan struct{}),
		createTime:  time.Now().Unix(),
	}

	// Verify it exists
	if _, ok := hub.cons[zone][id]; !ok {
		t.Fatal("Setup failed: connection not created")
	}

	// Unregister
	hub.UnRegisterBlock(zone, id)

	// Verify it's removed
	if _, ok := hub.cons[zone][id]; ok {
		t.Error("UnRegisterBlock() did not remove connection")
	}
}

func TestHub_SendMessage(t *testing.T) {
	hub := NewHub(nil)

	// Setup: create a zone with a connection
	zone := "test-zone"
	id := "test-id"
	hub.cons[zone] = make(map[string]Link)
	hub.cons[zone][id] = Link{
		messageChan: make(chan *Message, 10),
		allowPush:   make(chan struct{}, 10),
		createTime:  time.Now().Unix(),
	}

	tests := []struct {
		name    string
		pkg     Packet
		wantErr bool
	}{
		{
			name: "Send to specific client",
			pkg: Packet{
				Message: &Message{
					ID:    "msg1",
					Event: "test",
					Data:  "test data",
				},
				Zone:      zone,
				ClientID:  id,
				Broadcast: false,
			},
			wantErr: false,
		},
		{
			name: "Broadcast to zone",
			pkg: Packet{
				Message: &Message{
					ID:    "msg2",
					Event: "broadcast",
					Data:  "broadcast data",
				},
				Zone:      zone,
				ClientID:  "",
				Broadcast: true,
			},
			wantErr: false,
		},
		{
			name: "Send to non-existent zone",
			pkg: Packet{
				Message: &Message{
					ID:    "msg3",
					Event: "test",
					Data:  "data",
				},
				Zone:      "non-existent",
				ClientID:  "",
				Broadcast: true,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := hub.SendMessage(tt.pkg)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Mock log for testing
type mockLog struct{}

func (m *mockLog) Info(args ...interface{})                  {}
func (m *mockLog) Infoln(args ...interface{})                {}
func (m *mockLog) Infof(format string, args ...interface{})  {}
func (m *mockLog) Debug(args ...interface{})                 {}
func (m *mockLog) Debugln(args ...interface{})               {}
func (m *mockLog) Debugf(format string, args ...interface{}) {}
func (m *mockLog) Warn(args ...interface{})                  {}
func (m *mockLog) Warnln(args ...interface{})                {}
func (m *mockLog) Warnf(format string, args ...interface{})  {}
func (m *mockLog) Error(args ...interface{})                 {}
func (m *mockLog) Errorln(args ...interface{})               {}
func (m *mockLog) Errorf(format string, args ...interface{}) {}
