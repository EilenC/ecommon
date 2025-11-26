package sse

import (
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
