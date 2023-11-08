package sse

import (
	"fmt"
	"testing"
)

func TestMessage_Format(t *testing.T) {
	type fields struct {
		Event string
		Data  string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "success-text",
			fields: fields{
				Event: "info",
				Data:  "123abc",
			},
		},
		{
			name: "success-json",
			fields: fields{
				Event: "info",
				Data:  `{"name":"abc","age":16}`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Message{
				Event: tt.fields.Event,
				Data:  tt.fields.Data,
			}
			tt.want = fmt.Sprintf("event: %s\ndata: %s\n\n", m.Event, m.Data)
			if got := m.Format(); got != tt.want {
				t.Errorf("Format() = %v, want %v", got, tt.want)
			}
		})
	}
}
