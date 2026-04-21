package ecommon

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestGetAttachmentName(t *testing.T) {
	type args struct {
		path string
		sep  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Contains custom attachment names",
			args: args{
				path: "https://baike.seekhill.com/uploads/202106/1624354355xY7cLkuE_s.png|golang.png",
			},
			want: "golang.png",
		},
		{
			name: "Contains custom attachment names",
			args: args{
				path: "https://baike.seekhill.com/uploads/202106/1624354355xY7cLkuE_s.png-golang.png",
				sep:  "-",
			},
			want: "golang.png",
		},
		{
			name: "Does not include custom attachment names",
			args: args{
				path: "https://baike.seekhill.com/uploads/202106/1624354355xY7cLkuE_s.png",
			},
			want: "1624354355xY7cLkuE_s.png",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetAttachmentName(tt.args.path, tt.args.sep); got != tt.want {
				t.Errorf("GetAttachmentName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFileContent(t *testing.T) {
	// Create a temporary file for local file testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := []byte("test file content")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/missing" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("http content"))
	}))
	defer server.Close()

	tests := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{
			name:    "Local file",
			url:     testFile,
			want:    "test file content",
			wantErr: false,
		},
		{
			name:    "HTTP URL",
			url:     server.URL,
			want:    "http content",
			wantErr: false,
		},
		{
			name:    "HTTP error",
			url:     server.URL + "/missing",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Local path containing URL text",
			url:     filepath.Join(tempDir, "contains-http://-text.txt"),
			want:    "",
			wantErr: true,
		},
		{
			name:    "Non-existent local file",
			url:     "/non/existent/file.txt",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetFileContent(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFileContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && string(got) != tt.want {
				t.Errorf("GetFileContent() = %v, want %v", string(got), tt.want)
			}
		})
	}
}

func TestDownloadFile(t *testing.T) {
	// Create test servers
	successServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "12")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test content"))
	}))
	defer successServer.Close()

	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer errorServer.Close()

	shortBodyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "20")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("short"))
	}))
	defer shortBodyServer.Close()

	tests := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{
			name:    "Successful download",
			url:     successServer.URL,
			want:    "test content",
			wantErr: false,
		},
		{
			name:    "404 error",
			url:     errorServer.URL,
			want:    "",
			wantErr: true,
		},
		{
			name:    "Invalid URL",
			url:     "://bad-url",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Read error",
			url:     shortBodyServer.URL,
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DownloadFile(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("DownloadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != nil && string(*got) != tt.want {
				t.Errorf("DownloadFile() = %v, want %v", string(*got), tt.want)
			}
		})
	}
}

func TestCreateFile(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "Create file in existing directory",
			path:    filepath.Join(tempDir, "test1.txt"),
			wantErr: false,
		},
		{
			name:    "Create file with nested directories",
			path:    filepath.Join(tempDir, "nested", "dir", "test2.txt"),
			wantErr: false,
		},
		{
			name:    "Create file in deep nested path",
			path:    filepath.Join(tempDir, "a", "b", "c", "d", "test3.txt"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateFile(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil {
				defer got.Close()
				// Verify file was created
				if _, err := os.Stat(tt.path); os.IsNotExist(err) {
					t.Errorf("CreateFile() file was not created at %v", tt.path)
				}
			}
		})
	}
}

func TestCreateFileErrors(t *testing.T) {
	tempDir := t.TempDir()
	parentFile := filepath.Join(tempDir, "parent")
	if err := os.WriteFile(parentFile, []byte("not a dir"), 0644); err != nil {
		t.Fatalf("failed to create parent file: %v", err)
	}

	tests := []struct {
		name string
		path string
	}{
		{
			name: "mkdir fails",
			path: string([]byte{'b', 'a', 'd', 0, 'd', 'i', 'r', os.PathSeparator, 'x'}),
		},
		{
			name: "create fails when parent is file",
			path: filepath.Join(parentFile, "child.txt"),
		},
		{
			name: "create fails with empty file path",
			path: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if file, err := CreateFile(tt.path); err == nil {
				if file != nil {
					file.Close()
				}
				t.Fatalf("CreateFile() expected error")
			}
		})
	}
}
