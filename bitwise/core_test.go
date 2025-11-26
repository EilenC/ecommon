package bitwise

import (
	"bytes"
	"testing"
)

func Test_generateKey(t *testing.T) {
	tests := []struct {
		name string
		seed string
		want int // key length
	}{
		{
			name: "Generate key with seed 'test'",
			seed: "test",
			want: 32,
		},
		{
			name: "Generate key with seed 'password'",
			seed: "password",
			want: 32,
		},
		{
			name: "Generate key with empty seed",
			seed: "",
			want: 32,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateKey(tt.seed)
			if len(got) != tt.want {
				t.Errorf("generateKey() length = %v, want %v", len(got), tt.want)
			}
			// Test deterministic: same seed should produce same key
			got2 := generateKey(tt.seed)
			if !bytes.Equal(got, got2) {
				t.Errorf("generateKey() is not deterministic for seed %v", tt.seed)
			}
		})
	}
}

func Test_hashString(t *testing.T) {
	tests := []struct {
		name string
		s    string
	}{
		{
			name: "Hash 'test'",
			s:    "test",
		},
		{
			name: "Hash 'password'",
			s:    "password",
		},
		{
			name: "Hash empty string",
			s:    "",
		},
		{
			name: "Hash unicode",
			s:    "你好世界",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hashString(tt.s)
			// Test deterministic: same string should produce same hash
			got2 := hashString(tt.s)
			if got != got2 {
				t.Errorf("hashString() is not deterministic for string %v", tt.s)
			}
		})
	}
}

func Test_encrypt(t *testing.T) {
	key := []byte("testkey12345678901234567890123")
	tests := []struct {
		name string
		data []byte
		key  []byte
	}{
		{
			name: "Encrypt simple text",
			data: []byte("hello world"),
			key:  key,
		},
		{
			name: "Encrypt empty data",
			data: []byte(""),
			key:  key,
		},
		{
			name: "Encrypt binary data",
			data: []byte{0x00, 0x01, 0x02, 0x03, 0xFF},
			key:  key,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := encrypt(tt.data, tt.key)
			if len(got) != len(tt.data) {
				t.Errorf("encrypt() length = %v, want %v", len(got), len(tt.data))
			}
			// Encrypted data should be different from original (unless empty)
			if len(tt.data) > 0 && bytes.Equal(got, tt.data) {
				t.Errorf("encrypt() did not change the data")
			}
		})
	}
}

func Test_decrypt(t *testing.T) {
	key := []byte("testkey12345678901234567890123")
	tests := []struct {
		name string
		data []byte
		key  []byte
	}{
		{
			name: "Decrypt simple text",
			data: []byte("hello world"),
			key:  key,
		},
		{
			name: "Decrypt empty data",
			data: []byte(""),
			key:  key,
		},
		{
			name: "Decrypt binary data",
			data: []byte{0x00, 0x01, 0x02, 0x03, 0xFF},
			key:  key,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := decrypt(tt.data, tt.key)
			if len(got) != len(tt.data) {
				t.Errorf("decrypt() length = %v, want %v", len(got), len(tt.data))
			}
		})
	}
}

func Test_apply(t *testing.T) {
	key := []byte("testkey12345678901234567890123")
	tests := []struct {
		name string
		data []byte
		key  []byte
	}{
		{
			name: "Apply XOR to text",
			data: []byte("hello world"),
			key:  key,
		},
		{
			name: "Apply XOR to empty data",
			data: []byte(""),
			key:  key,
		},
		{
			name: "Apply XOR to binary",
			data: []byte{0x00, 0x01, 0x02, 0x03, 0xFF},
			key:  key,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := apply(tt.data, tt.key)
			if len(got) != len(tt.data) {
				t.Errorf("apply() length = %v, want %v", len(got), len(tt.data))
			}
			// Test XOR property: applying twice should return original
			got2 := apply(got, tt.key)
			if !bytes.Equal(got2, tt.data) {
				t.Errorf("apply() twice did not return original data")
			}
		})
	}
}

func Test_encryptDecryptRoundTrip(t *testing.T) {
	key := generateKey("test-seed")
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "Round trip simple text",
			data: []byte("hello world"),
		},
		{
			name: "Round trip long text",
			data: []byte("This is a longer text to test encryption and decryption with multiple bytes"),
		},
		{
			name: "Round trip binary data",
			data: []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted := encrypt(tt.data, key)
			decrypted := decrypt(encrypted, key)
			if !bytes.Equal(decrypted, tt.data) {
				t.Errorf("encrypt/decrypt round trip failed: got %v, want %v", decrypted, tt.data)
			}
		})
	}
}
