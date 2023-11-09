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
			gotMsg := got.String()
			if len(gotMsg) != len(tt.want) || gotMsg != tt.want {
				t.Errorf("Format() got = %s, want %s", gotMsg, tt.want)
			}
		})
	}
}
