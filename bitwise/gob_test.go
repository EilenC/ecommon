package bitwise

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetExt(t *testing.T) {
	tests := []struct {
		name   string
		newExt string
	}{
		{
			name:   "Set custom extension",
			newExt: ".custom",
		},
		{
			name:   "Set another extension",
			newExt: ".encrypted",
		},
		{
			name:   "Set extension without dot",
			newExt: "enc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalExt := ext
			defer func() { ext = originalExt }() // Restore original

			SetExt(tt.newExt)
			if ext != tt.newExt {
				t.Errorf("SetExt() ext = %v, want %v", ext, tt.newExt)
			}
		})
	}
}

func TestSaveFile(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name       string
		encrypted  []byte
		outputFile string
		wantErr    bool
	}{
		{
			name:       "Save valid encrypted data",
			encrypted:  []byte("encrypted data"),
			outputFile: filepath.Join(tempDir, "test1.txt"),
			wantErr:    false,
		},
		{
			name:       "Save empty encrypted data",
			encrypted:  []byte(""),
			outputFile: filepath.Join(tempDir, "test2.txt"),
			wantErr:    false,
		},
		{
			name:       "Save with nested directory",
			encrypted:  []byte("encrypted data"),
			outputFile: filepath.Join(tempDir, "nested", "dir", "test3.txt"),
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SaveFile(tt.encrypted, tt.outputFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Verify file was created with correct extension
				expectedPath := tt.outputFile + ext
				if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
					t.Errorf("SaveFile() file was not created at %v", expectedPath)
				}
				// Verify returned filename
				if got != filepath.Base(expectedPath) {
					t.Errorf("SaveFile() returned filename = %v, want %v", got, filepath.Base(expectedPath))
				}
				// Clean up
				os.Remove(expectedPath)
			}
		})
	}
}

func TestGetRealFileName(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     string
	}{
		{
			name:     "Get real filename with default extension",
			filePath: "/path/to/file.txt.bitwise",
			want:     "file.txt",
		},
		{
			name:     "Get real filename from basename",
			filePath: "file.doc.bitwise",
			want:     "file.doc",
		},
		{
			name:     "Get real filename with nested path",
			filePath: "/a/b/c/test.pdf.bitwise",
			want:     "test.pdf",
		},
		{
			name:     "Get real filename without extension",
			filePath: "/path/to/file",
			want:     "file",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetRealFileName(tt.filePath)
			if got != tt.want {
				t.Errorf("GetRealFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRealFileNameWithCustomExt(t *testing.T) {
	originalExt := ext
	defer func() { ext = originalExt }() // Restore original

	SetExt(".custom")

	tests := []struct {
		name     string
		filePath string
		want     string
	}{
		{
			name:     "Get real filename with custom extension",
			filePath: "/path/to/file.txt.custom",
			want:     "file.txt",
		},
		{
			name:     "Get real filename basename with custom ext",
			filePath: "document.pdf.custom",
			want:     "document.pdf",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetRealFileName(tt.filePath)
			if got != tt.want {
				t.Errorf("GetRealFileName() with custom ext = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSaveAndGetRealFileName(t *testing.T) {
	tempDir := t.TempDir()

	encrypted := []byte("test encrypted data")
	outputFile := filepath.Join(tempDir, "original.txt")

	savedName, err := SaveFile(encrypted, outputFile)
	if err != nil {
		t.Fatalf("SaveFile() failed: %v", err)
	}
	defer os.Remove(filepath.Join(tempDir, savedName))

	// Verify saved name has extension
	if !strings.HasSuffix(savedName, ext) {
		t.Errorf("SaveFile() returned name without extension: %v", savedName)
	}

	// Get real filename
	realName := GetRealFileName(filepath.Join(tempDir, savedName))
	expectedReal := "original.txt"

	if realName != expectedReal {
		t.Errorf("GetRealFileName() = %v, want %v", realName, expectedReal)
	}
}
