package sse

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		method         string
		reconnectDelay time.Duration
		wantDelay      time.Duration
	}{
		{
			name:           "Create client with custom delay",
			url:            "http://localhost:8080/events",
			method:         "GET",
			reconnectDelay: 5 * time.Second,
			wantDelay:      5 * time.Second,
		},
		{
			name:           "Create client with default delay",
			url:            "http://example.com/sse",
			method:         "POST",
			reconnectDelay: 0,
			wantDelay:      3 * time.Second,
		},
		{
			name:           "Create client with different URL",
			url:            "https://api.example.com/stream",
			method:         "GET",
			reconnectDelay: 10 * time.Second,
			wantDelay:      10 * time.Second,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewClient(tt.url, tt.method, tt.reconnectDelay)
			if got == nil {
				t.Error("NewClient() returned nil")
				return
			}
			if got.url != tt.url {
				t.Errorf("NewClient() url = %v, want %v", got.url, tt.url)
			}
			if got.method != tt.method {
				t.Errorf("NewClient() method = %v, want %v", got.method, tt.method)
			}
			if got.reconnectDelay != tt.wantDelay {
				t.Errorf("NewClient() reconnectDelay = %v, want %v", got.reconnectDelay, tt.wantDelay)
			}
			if got.eventCallbacks == nil {
				t.Error("NewClient() eventCallbacks is nil")
			}
			if got.client == nil {
				t.Error("NewClient() client is nil")
			}
			if got.stopSignal == nil {
				t.Error("NewClient() stopSignal is nil")
			}
			if got.exitSignal == nil {
				t.Error("NewClient() exitSignal is nil")
			}
		})
	}
}

func TestClient_SubscribeEvent(t *testing.T) {
	client := NewClient("http://localhost/events", "GET", 3*time.Second)

	callbackCalled := false
	callback := func(message *Message) {
		callbackCalled = true
	}

	eventName := "test-event"
	client.SubscribeEvent(eventName, callback)

	if client.eventCallbacks[eventName] == nil {
		t.Error("SubscribeEvent() callback was not registered")
		return
	}

	// Test callback can be called
	client.eventCallbacks[eventName](&Message{Event: eventName})
	if !callbackCalled {
		t.Error("SubscribeEvent() callback was not called")
	}
}

func TestClient_SubscribeMultipleEvents(t *testing.T) {
	client := NewClient("http://localhost/events", "GET", 3*time.Second)

	event1Called := false
	event2Called := false

	client.SubscribeEvent("event1", func(m *Message) {
		event1Called = true
	})

	client.SubscribeEvent("event2", func(m *Message) {
		event2Called = true
	})

	// Call event1
	if cb, ok := client.eventCallbacks["event1"]; ok {
		cb(&Message{Event: "event1"})
	}

	if !event1Called {
		t.Error("event1 callback was not called")
	}
	if event2Called {
		t.Error("event2 callback should not have been called")
	}

	// Call event2
	if cb, ok := client.eventCallbacks["event2"]; ok {
		cb(&Message{Event: "event2"})
	}

	if !event2Called {
		t.Error("event2 callback was not called")
	}
}

func TestClient_OnConnection(t *testing.T) {
	client := NewClient("http://localhost/events", "GET", 3*time.Second)

	handlerCalled := false
	handler := func() {
		handlerCalled = true
	}

	client.OnConnection(handler)

	if client.connectionHandler == nil {
		t.Error("OnConnection() handler was not set")
		return
	}

	// Test handler can be called
	client.connectionHandler()
	if !handlerCalled {
		t.Error("OnConnection() handler was not called")
	}
}

func TestClient_OnDisconnect(t *testing.T) {
	client := NewClient("http://localhost/events", "GET", 3*time.Second)

	var receivedErr string
	handler := func(err string) {
		receivedErr = err
	}

	client.OnDisconnect(handler)

	if client.disconnectHandler == nil {
		t.Error("OnDisconnect() handler was not set")
		return
	}

	// Test handler can be called
	testErr := "test error"
	client.disconnectHandler(testErr)
	if receivedErr != testErr {
		t.Errorf("OnDisconnect() received error = %v, want %v", receivedErr, testErr)
	}
}

func TestClient_OnExit(t *testing.T) {
	client := NewClient("http://localhost/events", "GET", 3*time.Second)

	handlerCalled := false
	handler := func() {
		handlerCalled = true
	}

	client.OnExit(handler)

	if client.exitHandler == nil {
		t.Error("OnExit() handler was not set")
		return
	}

	// Test handler can be called
	client.exitHandler()
	if !handlerCalled {
		t.Error("OnExit() handler was not called")
	}
}

func TestClient_Stop(t *testing.T) {
	client := NewClient("http://localhost/events", "GET", 3*time.Second)

	// Test that Stop sends to exitSignal
	done := make(chan bool, 1)
	go func() {
		client.Stop()
		done <- true
	}()

	select {
	case <-client.exitSignal:
		// Success - signal was sent
	case <-time.After(1 * time.Second):
		t.Error("Stop() did not send to exitSignal")
	}

	<-done
}

func TestClient_Start(t *testing.T) {
	client := NewClient("://bad-url", "GET", time.Millisecond)
	done := make(chan struct{})
	client.OnExit(func() {
		close(done)
	})

	go client.Start()
	client.Stop()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Start() did not call exit handler")
	}
}

func TestClient_connectInvalidRequest(t *testing.T) {
	client := NewClient("http://localhost/events", "\n", time.Millisecond)
	done := make(chan struct{})
	go func() {
		client.connect()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("connect() did not return for invalid request")
	}
}

func TestClient_connectDoError(t *testing.T) {
	client := NewClient("http://localhost/events", http.MethodGet, time.Millisecond)
	client.client = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("connect failed")
	})}

	done := make(chan struct{})
	go func() {
		client.connect()
		close(done)
	}()
	time.Sleep(5 * time.Millisecond)
	client.Stop()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("connect() did not stop after transport error")
	}
}

func TestClient_connectDoErrorStoppedBeforeRetry(t *testing.T) {
	requested := make(chan struct{}, 1)
	client := NewClient("http://localhost/events", http.MethodGet, time.Hour)
	client.client = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		requested <- struct{}{}
		return nil, errors.New("connect failed")
	})}

	done := make(chan struct{})
	go func() {
		client.connect()
		close(done)
	}()

	select {
	case <-requested:
	case <-time.After(time.Second):
		t.Fatal("connect() did not attempt request")
	}
	client.Stop()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("connect() did not stop while waiting to retry")
	}
}

func TestClient_connectStatusError(t *testing.T) {
	requested := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requested <- struct{}{}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, http.MethodGet, time.Millisecond)
	done := make(chan struct{})
	go func() {
		client.connect()
		close(done)
	}()

	select {
	case <-requested:
	case <-time.After(time.Second):
		t.Fatal("connect() did not request server")
	}
	client.Stop()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("connect() did not stop after status error")
	}
}

func TestClient_connectStatusErrorRetries(t *testing.T) {
	requests := make(chan struct{}, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- struct{}{}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, http.MethodGet, time.Millisecond)
	done := make(chan struct{})
	go func() {
		client.connect()
		close(done)
	}()

	for i := 0; i < 2; i++ {
		select {
		case <-requests:
		case <-time.After(time.Second):
			t.Fatalf("connect() request %d timed out", i+1)
		}
	}
	client.Stop()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("connect() did not stop after retry")
	}
}

func TestClient_connectStopsWhileConnected(t *testing.T) {
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		<-release
	}))
	defer server.Close()
	defer close(release)

	client := NewClient(server.URL, http.MethodGet, time.Millisecond)
	done := make(chan struct{})
	go func() {
		client.connect()
		close(done)
	}()

	time.Sleep(10 * time.Millisecond)
	client.Stop()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("connect() did not stop while connected")
	}
}

func TestClient_connectAlreadyStopped(t *testing.T) {
	client := NewClient("http://localhost/events", http.MethodGet, time.Millisecond)
	client.Stop()
	client.connect()
}

func TestClient_StartWithoutExitHandler(t *testing.T) {
	client := NewClient("://bad-url", "GET", time.Millisecond)
	done := make(chan struct{})
	go func() {
		client.Start()
		close(done)
	}()
	client.Stop()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Start() without exit handler did not return")
	}
}

func TestClient_listenEventsClosedConnectionError(t *testing.T) {
	client := NewClient("http://localhost/events", http.MethodGet, time.Millisecond)
	client.listenEvents(errorReadCloser{err: errors.New("use of closed network connection")})
}

func TestClient_connectSuccessAndListenEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("event: test\ndata: hello\n\n"))
	}))
	defer server.Close()

	client := NewClient(server.URL, http.MethodGet, time.Millisecond)
	connected := make(chan struct{}, 1)
	disconnected := make(chan string, 1)
	event := make(chan *Message, 1)
	client.OnConnection(func() {
		connected <- struct{}{}
	})
	client.OnDisconnect(func(err string) {
		disconnected <- err
	})
	client.SubscribeEvent("test", func(message *Message) {
		event <- message
	})

	done := make(chan struct{})
	go func() {
		client.connect()
		close(done)
	}()

	select {
	case <-connected:
	case <-time.After(time.Second):
		t.Fatal("connect() did not call connection handler")
	}
	select {
	case msg := <-event:
		if msg.Data != "hello\n" {
			t.Fatalf("event data = %q, want %q", msg.Data, "hello\n")
		}
	case <-time.After(time.Second):
		t.Fatal("listenEvents() did not call event callback")
	}
	select {
	case got := <-disconnected:
		if got != "client stopped" {
			t.Fatalf("disconnect err = %q, want client stopped", got)
		}
	case <-time.After(time.Second):
		t.Fatal("connect() did not call disconnect handler")
	}
	client.Stop()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("connect() did not stop")
	}
}

func TestClient_listenEventsEOF(t *testing.T) {
	client := NewClient("http://localhost/events", http.MethodGet, time.Millisecond)
	client.listenEvents(io.NopCloser(strings.NewReader("")))

	select {
	case <-client.stopSignal:
	default:
		t.Fatal("listenEvents() did not signal stop on EOF")
	}
}

func TestClient_AllHandlers(t *testing.T) {
	client := NewClient("http://localhost/events", "GET", 3*time.Second)

	connectionCalled := false
	disconnectCalled := false
	exitCalled := false
	eventCalled := false

	client.OnConnection(func() {
		connectionCalled = true
	})

	client.OnDisconnect(func(err string) {
		disconnectCalled = true
	})

	client.OnExit(func() {
		exitCalled = true
	})

	client.SubscribeEvent("test", func(m *Message) {
		eventCalled = true
	})

	// Verify all handlers are set
	if client.connectionHandler == nil {
		t.Error("connectionHandler not set")
	}
	if client.disconnectHandler == nil {
		t.Error("disconnectHandler not set")
	}
	if client.exitHandler == nil {
		t.Error("exitHandler not set")
	}
	if client.eventCallbacks["test"] == nil {
		t.Error("event callback not set")
	}

	// Call all handlers
	client.connectionHandler()
	client.disconnectHandler("error")
	client.exitHandler()
	client.eventCallbacks["test"](&Message{})

	if !connectionCalled || !disconnectCalled || !exitCalled || !eventCalled {
		t.Error("Not all handlers were called successfully")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

type errorReadCloser struct {
	err error
}

func (r errorReadCloser) Read([]byte) (int, error) {
	return 0, r.err
}

func (r errorReadCloser) Close() error {
	return nil
}
