package sse

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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

	hub.UnRegisterBlock(zone, "missing")
	hub.UnRegisterBlock("missing-zone", id)
}

func TestHub_deferStartBroadcast(t *testing.T) {
	hub := NewHub(&mockLog{})
	hub.deferStartBroadcast()

	func() {
		defer hub.deferStartBroadcast()
		panic("boom")
	}()
}

func TestHub_broadcastMessageAndReply(t *testing.T) {
	hub := NewHub(&mockLog{})
	message := &Message{Event: "event", Data: "data"}
	hub.cons["zone-a"] = map[string]Link{
		"client-a": {messageChan: make(chan *Message, 1), allowPush: make(chan struct{}, 1)},
	}
	hub.cons["zone-b"] = map[string]Link{
		"client-b": {messageChan: make(chan *Message), allowPush: make(chan struct{}, 1)},
	}

	hub.broadcastMessage(Packet{Message: message})

	select {
	case got := <-hub.cons["zone-a"]["client-a"].messageChan:
		if got != message {
			t.Fatalf("broadcastMessage() message = %v, want %v", got, message)
		}
	default:
		t.Fatal("broadcastMessage() did not send to buffered client")
	}
	hub.broadcastReply("zone-a", "client-a", message)
}

func TestHub_RegisterBlock(t *testing.T) {
	hub := NewHub(nil)
	connected := make(chan string, 1)
	disconnected := make(chan string, 1)
	hub.ConnectedFunc = func(clientID string) {
		connected <- clientID
	}
	hub.DisconnectFunc = func(clientID string) {
		disconnected <- clientID
	}

	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/sse", nil).WithContext(ctx)
	recorder := httptest.NewRecorder()
	done := make(chan struct{})

	go func() {
		hub.RegisterBlock(recorder, req, "", func() string { return "client-1" })
		close(done)
	}()

	select {
	case id := <-connected:
		if id != "client-1" {
			t.Fatalf("ConnectedFunc id = %v, want client-1", id)
		}
	case <-time.After(time.Second):
		t.Fatal("RegisterBlock() connection timed out")
	}
	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("RegisterBlock() did not return after context cancellation")
	}
	select {
	case id := <-disconnected:
		if id != "client-1" {
			t.Fatalf("DisconnectFunc id = %v, want client-1", id)
		}
	case <-time.After(time.Second):
		t.Fatal("DisconnectFunc was not called")
	}
	if !strings.Contains(recorder.Body.String(), "Connection Successful!") {
		t.Fatalf("RegisterBlock() body = %q", recorder.Body.String())
	}
	if recorder.Header().Get("Content-Type") != "text/event-stream; charset=utf-8" {
		t.Fatalf("Content-Type = %q", recorder.Header().Get("Content-Type"))
	}
}

func TestHub_RegisterBlockUnsupported(t *testing.T) {
	hub := NewHub(nil)
	writer := &plainResponseWriter{header: http.Header{}}
	req := httptest.NewRequest(http.MethodGet, "/sse", nil)

	hub.RegisterBlock(writer, req, "zone", func() string { return "id" })

	if writer.status != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", writer.status, http.StatusInternalServerError)
	}
}

func TestHub_RegisterBlockDefaultUUID(t *testing.T) {
	hub := NewHub(nil)
	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/sse", nil).WithContext(ctx)
	recorder := httptest.NewRecorder()
	done := make(chan struct{})

	go func() {
		hub.RegisterBlock(recorder, req, "zone", nil)
		close(done)
	}()

	deadline := time.After(time.Second)
	for {
		hub.block.Lock()
		ready := len(hub.cons["zone"]) == 1
		hub.block.Unlock()
		if ready {
			break
		}
		select {
		case <-deadline:
			t.Fatal("RegisterBlock() did not register default UUID client")
		default:
			time.Sleep(time.Millisecond)
		}
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("RegisterBlock() with default UUID did not stop")
	}
}

func TestHub_RegisterBlockWriteError(t *testing.T) {
	hub := NewHub(&mockLog{})
	req := httptest.NewRequest(http.MethodGet, "/sse", nil)
	writer := &flushErrorResponseWriter{errorResponseWriter: errorResponseWriter{}}

	hub.RegisterBlock(writer, req, "zone", func() string { return "id" })
}

func TestMessage_WriteConnect(t *testing.T) {
	recorder := httptest.NewRecorder()
	if err := (&Message{ID: "1", Event: "event", Data: "data"}).WriteConnect(recorder); err != nil {
		t.Fatalf("WriteConnect() error = %v", err)
	}
	if !strings.Contains(recorder.Body.String(), "data: data") {
		t.Fatalf("WriteConnect() body = %q", recorder.Body.String())
	}
	if err := (&Message{}).WriteConnect(recorder); err == nil {
		t.Fatal("WriteConnect() expected empty message error")
	}
	if err := (&Message{Data: "data"}).WriteConnect(&errorResponseWriter{}); err == nil {
		t.Fatal("WriteConnect() expected writer error")
	}
}

func TestHub_SendMessage(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*Hub)
		pkg     Packet
		wantErr bool
	}{
		{
			name: "Send to specific client",
			setup: func(hub *Hub) {
				hub.cons["test-zone"] = map[string]Link{
					"test-id": {
						messageChan: make(chan *Message, 10),
						allowPush:   make(chan struct{}, 10),
						createTime:  time.Now().Unix(),
					},
				}
			},
			pkg: Packet{
				Message: &Message{
					ID:    "msg1",
					Event: "test",
					Data:  "test data",
				},
				Zone:      "test-zone",
				ClientID:  "test-id",
				Broadcast: false,
			},
			wantErr: false,
		},
		{
			name: "Broadcast to zone",
			setup: func(hub *Hub) {
				hub.cons["test-zone"] = map[string]Link{
					"test-id": {
						messageChan: make(chan *Message, 10),
						allowPush:   make(chan struct{}, 10),
						createTime:  time.Now().Unix(),
					},
				}
			},
			pkg: Packet{
				Message: &Message{
					ID:    "msg2",
					Event: "broadcast",
					Data:  "broadcast data",
				},
				Zone:      "test-zone",
				ClientID:  "",
				Broadcast: true,
			},
			wantErr: false,
		},
		{
			name: "Broadcast to all",
			setup: func(hub *Hub) {
				hub.cons["test-zone"] = map[string]Link{
					"test-id": {
						messageChan: make(chan *Message, 10),
						allowPush:   make(chan struct{}, 10),
						createTime:  time.Now().Unix(),
					},
				}
			},
			pkg: Packet{
				Message:   &Message{ID: "msg-all", Event: "broadcast", Data: "all"},
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
		{
			name: "Send to empty zone",
			setup: func(hub *Hub) {
				hub.cons["empty-zone"] = map[string]Link{}
			},
			pkg: Packet{
				Message:   &Message{ID: "msg4", Event: "test", Data: "data"},
				Zone:      "empty-zone",
				Broadcast: true,
			},
			wantErr: true,
		},
		{
			name: "Send to missing client",
			setup: func(hub *Hub) {
				hub.cons["test-zone"] = map[string]Link{
					"other": {
						messageChan: make(chan *Message, 1),
						allowPush:   make(chan struct{}, 1),
						createTime:  time.Now().Unix(),
					},
				}
			},
			pkg: Packet{
				Message:  &Message{ID: "msg5", Event: "test", Data: "data"},
				Zone:     "test-zone",
				ClientID: "missing",
			},
			wantErr: false,
		},
		{
			name: "Send to blocked client",
			setup: func(hub *Hub) {
				hub.cons["test-zone"] = map[string]Link{
					"test-id": {
						messageChan: make(chan *Message),
						allowPush:   make(chan struct{}),
						createTime:  time.Now().Unix(),
					},
				}
			},
			pkg: Packet{
				Message:  &Message{ID: "msg-blocked", Event: "test", Data: "data"},
				Zone:     "test-zone",
				ClientID: "test-id",
			},
			wantErr: true,
		},
		{
			name: "Send to specific client after allow push",
			setup: func(hub *Hub) {
				allowPush := make(chan struct{}, 1)
				allowPush <- struct{}{}
				hub.cons["test-zone"] = map[string]Link{
					"test-id": {
						messageChan: make(chan *Message, 10),
						allowPush:   allowPush,
						createTime:  time.Now().Unix(),
					},
				}
			},
			pkg: Packet{
				Message:  &Message{ID: "msg6", Event: "test", Data: "data"},
				Zone:     "test-zone",
				ClientID: "test-id",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hub := NewHub(nil)
			if tt.setup != nil {
				tt.setup(hub)
			}
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

type plainResponseWriter struct {
	header http.Header
	status int
	body   strings.Builder
}

func (w *plainResponseWriter) Header() http.Header {
	return w.header
}

func (w *plainResponseWriter) Write(b []byte) (int, error) {
	return w.body.Write(b)
}

func (w *plainResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
}

type errorResponseWriter struct{}

func (w *errorResponseWriter) Header() http.Header {
	return http.Header{}
}

func (w *errorResponseWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func (w *errorResponseWriter) WriteHeader(int) {}

type flushErrorResponseWriter struct {
	errorResponseWriter
}

func (w *flushErrorResponseWriter) Flush() {}
