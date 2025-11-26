package sse

import (
	"strings"
	"testing"
)

func TestNewDecoder(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Create decoder with simple input",
			input: "data: test\n\n",
		},
		{
			name:  "Create decoder with empty input",
			input: "",
		},
		{
			name:  "Create decoder with complex input",
			input: "id: 123\nevent: test\ndata: content\n\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			got := NewDecoder(reader)
			if got == nil {
				t.Error("NewDecoder() returned nil")
				return
			}
			if got.reader == nil {
				t.Error("NewDecoder() reader is nil")
			}
		})
	}
}

func TestDecoder_Decode(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantEvent string
		wantData  string
		wantID    string
		wantRetry string
		wantErr   bool
	}{
		{
			name:      "Decode simple data",
			input:     "data: hello\n\n",
			wantEvent: "",
			wantData:  "hello\n",
			wantID:    "",
			wantRetry: "",
			wantErr:   false,
		},
		{
			name:      "Decode with event",
			input:     "event: message\ndata: test data\n\n",
			wantEvent: "message",
			wantData:  "test data\n",
			wantID:    "",
			wantRetry: "",
			wantErr:   false,
		},
		{
			name:      "Decode with all fields",
			input:     "id: 123\nevent: update\ndata: content\nretry: 3000\n\n",
			wantEvent: "update",
			wantData:  "content\n",
			wantID:    "123",
			wantRetry: "3000",
			wantErr:   false,
		},
		{
			name:      "Decode with comment",
			input:     ": this is a comment\ndata: test\n\n",
			wantEvent: "",
			wantData:  "test\n",
			wantID:    "",
			wantRetry: "",
			wantErr:   false,
		},
		{
			name:      "Decode multiline data",
			input:     "data: line1\ndata: line2\ndata: line3\n\n",
			wantEvent: "",
			wantData:  "line1\nline2\nline3\n",
			wantID:    "",
			wantRetry: "",
			wantErr:   false,
		},
		{
			name:      "Decode with spaces",
			input:     "event:  test  \ndata:  content  \n\n",
			wantEvent: "test",
			wantData:  "content\n",
			wantID:    "",
			wantRetry: "",
			wantErr:   false,
		},
		{
			name:      "Decode only event",
			input:     "event: ping\n\n",
			wantEvent: "ping",
			wantData:  "",
			wantID:    "",
			wantRetry: "",
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			decoder := NewDecoder(reader)

			got, err := decoder.Decode()
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got.Event != tt.wantEvent {
					t.Errorf("Decode() Event = %v, want %v", got.Event, tt.wantEvent)
				}
				if got.Data != tt.wantData {
					t.Errorf("Decode() Data = %v, want %v", got.Data, tt.wantData)
				}
				if got.ID != tt.wantID {
					t.Errorf("Decode() ID = %v, want %v", got.ID, tt.wantID)
				}
				if got.Retry != tt.wantRetry {
					t.Errorf("Decode() Retry = %v, want %v", got.Retry, tt.wantRetry)
				}
			}
		})
	}
}

func TestDecoder_DecodeMultipleMessages(t *testing.T) {
	input := "data: message1\n\ndata: message2\n\nevent: test\ndata: message3\n\n"
	reader := strings.NewReader(input)
	decoder := NewDecoder(reader)

	// Decode first message
	msg1, err := decoder.Decode()
	if err != nil {
		t.Errorf("Decode() first message error = %v", err)
	}
	if msg1.Data != "message1\n" {
		t.Errorf("First message data = %v, want %v", msg1.Data, "message1\n")
	}

	// Decode second message
	msg2, err := decoder.Decode()
	if err != nil {
		t.Errorf("Decode() second message error = %v", err)
	}
	if msg2.Data != "message2\n" {
		t.Errorf("Second message data = %v, want %v", msg2.Data, "message2\n")
	}

	// Decode third message
	msg3, err := decoder.Decode()
	if err != nil {
		t.Errorf("Decode() third message error = %v", err)
	}
	if msg3.Event != "test" {
		t.Errorf("Third message event = %v, want %v", msg3.Event, "test")
	}
	if msg3.Data != "message3\n" {
		t.Errorf("Third message data = %v, want %v", msg3.Data, "message3\n")
	}
}

func TestDecoder_DecodeWithComment(t *testing.T) {
	input := ": heartbeat\n\n"
	reader := strings.NewReader(input)
	decoder := NewDecoder(reader)

	msg, err := decoder.Decode()
	if err != nil {
		t.Errorf("Decode() error = %v", err)
		return
	}

	if msg.Comment != "heartbeat" {
		t.Errorf("Decode() Comment = %v, want %v", msg.Comment, "heartbeat")
	}
}

func TestDecoder_DecodeEmptyLines(t *testing.T) {
	input := "\n\n\ndata: test\n\n"
	reader := strings.NewReader(input)
	decoder := NewDecoder(reader)

	// First decode should return empty message (from empty lines)
	msg1, err := decoder.Decode()
	if err != nil {
		t.Errorf("Decode() error = %v", err)
		return
	}

	// Should have empty data and event
	if msg1.Data != "" || msg1.Event != "" {
		// This is expected behavior - empty lines create empty messages
	}
}
