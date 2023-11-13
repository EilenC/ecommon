package sse

import (
	"bufio"
	"io"
	"strings"
)

// NewDecoder sever-sent events
func NewDecoder(reader io.Reader) *Decoder {
	return &Decoder{
		reader: bufio.NewReader(reader),
	}
}

// Decode sever-sent events message decode
func (d *Decoder) Decode() (*Message, error) {
	event := &Message{}

	for {
		line, err := d.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = strings.TrimSpace(line)

		if line == "" {
			return event, nil
		}

		if strings.HasPrefix(line, "event:") {
			event.Event = strings.TrimSpace(line[6:])
		} else if strings.HasPrefix(line, "data:") {
			event.Data += strings.TrimSpace(line[5:]) + "\n"
		} else if strings.HasPrefix(line, "id:") {
			event.ID = strings.TrimSpace(line[3:])
		} else if strings.HasPrefix(line, "retry:") {
			event.Retry = strings.TrimSpace(line[6:])
		} else if strings.HasPrefix(line, ":") {
			event.Comment = strings.TrimSpace(line[1:])
		}
	}
}
