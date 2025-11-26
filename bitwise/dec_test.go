package bitwise

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestDecrypt(t *testing.T) {
	tests := []struct {
		name    string
		seed    string
		wantErr bool
	}{
		{
			name:    "Decrypt with valid seed",
			seed:    "valid-seed",
			wantErr: false,
		},
		{
			name:    "Decrypt with empty seed",
			seed:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First encrypt some data
			originalData := []byte("test data for decryption")
			var encrypted []byte
			var err error

			if tt.seed != "" {
				encrypted, err = Encrypt(originalData, tt.seed)
				if err != nil {
					t.Fatalf("Failed to encrypt test data: %v", err)
				}
			} else {
				encrypted = []byte("dummy")
			}

			got, err := Decrypt(encrypted, tt.seed)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !bytes.Equal(got, originalData) {
				t.Errorf("Decrypt() = %v, want %v", got, originalData)
			}
		})
	}
}

func TestDecryptFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create and encrypt a test file
	testContent := []byte("test file content for decryption")
	testFile := "test.txt"
	testPath := filepath.Join(tempDir, testFile)

	if err := os.WriteFile(testPath, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	seed := "test-seed"
	encryptedFileName, err := EncryptFile(testFile, seed, tempDir)
	if err != nil {
		t.Fatalf("Failed to encrypt test file: %v", err)
	}

	encryptedPath := filepath.Join(tempDir, encryptedFileName)

	tests := []struct {
		name       string
		filePath   string
		seed       string
		filePrefix string
		wantErr    bool
	}{
		{
			name:       "Decrypt valid file",
			filePath:   encryptedPath,
			seed:       seed,
			filePrefix: "decrypted_",
			wantErr:    false,
		},
		{
			name:       "Decrypt with wrong seed",
			filePath:   encryptedPath,
			seed:       "wrong-seed",
			filePrefix: "decrypted_",
			wantErr:    false, // Will succeed but data will be wrong
		},
		{
			name:       "Decrypt with empty seed",
			filePath:   encryptedPath,
			seed:       "",
			filePrefix: "decrypted_",
			wantErr:    true,
		},
		{
			name:       "Decrypt non-existent file",
			filePath:   filepath.Join(tempDir, "nonexistent.bitwise"),
			seed:       seed,
			filePrefix: "decrypted_",
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecryptFile(tt.filePath, tt.seed, tt.filePrefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecryptFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != "" {
				// Verify decrypted file was created
				decryptedPath := filepath.Join(filepath.Dir(tt.filePath), got)
				if _, err := os.Stat(decryptedPath); os.IsNotExist(err) {
					t.Errorf("DecryptFile() decrypted file was not created")
				} else {
					// Clean up
					os.Remove(decryptedPath)
				}
			}
		})
	}

	// Clean up encrypted file
	os.Remove(encryptedPath)
}

func TestEncryptDecryptFileRoundTrip(t *testing.T) {
	tempDir := t.TempDir()

	testContent := []byte("Round trip test content with special chars: 你好世界 🌍")
	testFile := "roundtrip.txt"
	testPath := filepath.Join(tempDir, testFile)

	if err := os.WriteFile(testPath, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	seed := "roundtrip-seed"

	// Encrypt
	encryptedFileName, err := EncryptFile(testFile, seed, tempDir)
	if err != nil {
		t.Fatalf("EncryptFile() failed: %v", err)
	}
	defer os.Remove(filepath.Join(tempDir, encryptedFileName))

	// Decrypt
	decryptedFileName, err := DecryptFile(filepath.Join(tempDir, encryptedFileName), seed, "dec_")
	if err != nil {
		t.Fatalf("DecryptFile() failed: %v", err)
	}

	decryptedPath := filepath.Join(tempDir, decryptedFileName)
	defer os.Remove(decryptedPath)

	// Verify content
	decryptedContent, err := os.ReadFile(decryptedPath)
	if err != nil {
		t.Fatalf("Failed to read decrypted file: %v", err)
	}

	if !bytes.Equal(decryptedContent, testContent) {
		t.Errorf("Round trip failed: got %v, want %v", decryptedContent, testContent)
	}
}
