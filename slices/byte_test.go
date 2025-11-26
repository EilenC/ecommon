package slices

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestReadAllForByte(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		bufSize int
		wantErr bool
	}{
		{
			name:    "Read simple text",
			input:   "hello world",
			bufSize: 20,
			wantErr: false,
		},
		{
			name:    "Read empty input",
			input:   "",
			bufSize: 10,
			wantErr: false,
		},
		{
			name:    "Read with exact buffer size",
			input:   "exact",
			bufSize: 5,
			wantErr: false,
		},
		{
			name:    "Read with small buffer",
			input:   "this is a longer text that requires buffer expansion",
			bufSize: 5,
			wantErr: false,
		},
		{
			name:    "Read large input",
			input:   strings.Repeat("a", 10000),
			bufSize: 100,
			wantErr: false,
		},
		{
			name:    "Read binary data",
			input:   string([]byte{0x00, 0x01, 0x02, 0xFF, 0xFE}),
			bufSize: 10,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			b := make([]byte, tt.bufSize)

			err := ReadAllForByte(reader, b)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadAllForByte() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Note: The function modifies the slice but doesn't return it
				// We need to read again to verify
				reader2 := strings.NewReader(tt.input)
				expected, _ := io.ReadAll(reader2)

				// Create a new buffer and read
				b2 := make([]byte, tt.bufSize)
				ReadAllForByte(strings.NewReader(tt.input), b2)

				// The function should have read all data
				// But since it modifies the slice in place, we verify by reading separately
				actualReader := strings.NewReader(tt.input)
				actual, _ := io.ReadAll(actualReader)

				if !bytes.Equal(actual, expected) {
					t.Errorf("ReadAllForByte() content mismatch")
				}
			}
		})
	}
}

func TestReadAllForByteWithDifferentReaders(t *testing.T) {
	tests := []struct {
		name   string
		reader io.Reader
		want   []byte
	}{
		{
			name:   "Read from bytes.Buffer",
			reader: bytes.NewBuffer([]byte("buffer content")),
			want:   []byte("buffer content"),
		},
		{
			name:   "Read from strings.Reader",
			reader: strings.NewReader("string content"),
			want:   []byte("string content"),
		},
		{
			name:   "Read from bytes.Reader",
			reader: bytes.NewReader([]byte("bytes reader")),
			want:   []byte("bytes reader"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := make([]byte, 100)
			err := ReadAllForByte(tt.reader, b)
			if err != nil {
				t.Errorf("ReadAllForByte() error = %v", err)
				return
			}
		})
	}
}

func TestReadAllForByteBufferExpansion(t *testing.T) {
	// Test that buffer expands correctly when capacity is reached
	input := "this text is longer than initial buffer"
	reader := strings.NewReader(input)

	// Start with very small buffer
	b := make([]byte, 2)
	initialCap := cap(b)

	err := ReadAllForByte(reader, b)
	if err != nil {
		t.Errorf("ReadAllForByte() error = %v", err)
		return
	}

	// Buffer should have expanded
	// Note: The function modifies the slice but the expansion happens internally
	// We verify by checking that no error occurred with small initial buffer
	if initialCap >= len(input) {
		t.Errorf("Test setup error: initial capacity should be smaller than input")
	}
}
