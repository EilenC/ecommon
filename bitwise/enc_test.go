package bitwise

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEncrypt(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		seed    string
		wantErr bool
	}{
		{
			name:    "Encrypt with valid seed",
			data:    []byte("test data"),
			seed:    "valid-seed",
			wantErr: false,
		},
		{
			name:    "Encrypt with empty seed",
			data:    []byte("test data"),
			seed:    "",
			wantErr: true,
		},
		{
			name:    "Encrypt empty data",
			data:    []byte(""),
			seed:    "valid-seed",
			wantErr: false,
		},
		{
			name:    "Encrypt large data",
			data:    make([]byte, 10000),
			seed:    "valid-seed",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Encrypt(tt.data, tt.seed)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.data) {
					t.Errorf("Encrypt() length = %v, want %v", len(got), len(tt.data))
				}
			}
		})
	}
}

func TestEncryptFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test file
	testFile := "test.txt"
	testContent := []byte("test file content for encryption")
	testPath := filepath.Join(tempDir, testFile)
	if err := os.WriteFile(testPath, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		filePath string
		seed     string
		prePath  string
		wantErr  bool
	}{
		{
			name:     "Encrypt valid file",
			filePath: testFile,
			seed:     "test-seed",
			prePath:  tempDir,
			wantErr:  false,
		},
		{
			name:     "Encrypt with empty seed",
			filePath: testFile,
			seed:     "",
			prePath:  tempDir,
			wantErr:  true,
		},
		{
			name:     "Encrypt non-existent file",
			filePath: "nonexistent.txt",
			seed:     "test-seed",
			prePath:  tempDir,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EncryptFile(tt.filePath, tt.seed, tt.prePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncryptFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != "" {
				// Verify encrypted file was created - got contains the filename
				encryptedPath := filepath.Join(tt.prePath, got)
				if _, err := os.Stat(encryptedPath); os.IsNotExist(err) {
					t.Errorf("EncryptFile() encrypted file was not created at %v", encryptedPath)
				} else {
					// Clean up
					os.Remove(encryptedPath)
				}
			}
		})
	}
}
