package ecommon

import (
	"crypto/tls"
	"fmt"
	"github.com/EilenC/ecommon/slices"
	"net/http"
	"os"
	"strings"
)

// GetFileContent to retrieve file content
// Determine whether to use HTTP download by checking if the URL contains http://| https://
func GetFileContent(url string) ([]byte, error) {
	if strings.Contains(url, "http://") || strings.Contains(url, "https://") {
		b, err := DownloadFile(url)
		if err != nil {
			return nil, err
		}
		return *b, nil
	}
	b, readErr := os.ReadFile(url)
	if readErr != nil {
		return nil, fmt.Errorf("reading local send object file, please check Path:[%s]", url)
	}
	return b, nil
}

// GetAttachmentName determines the file name in the path using a custom delimiter
// (if there is no custom delimiter, the last part of the path is taken by default)
// Path file path
// Sep custom delimiter defaults to`|`
func GetAttachmentName(path, sep string) string {
	if sep == "" {
		sep = "|"
	}
	names := strings.Split(path, sep)
	if len(names) == 2 {
		return names[1]
	}
	names = strings.Split(path, "/")
	return names[len(names)-1]
}

// DownloadFile using HTTP get and skip SSL verification, returning a pointer to [] byte
func DownloadFile(url string) (*[]byte, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK || resp.Body == nil {
		return nil, err
	}
	//resource length from response header and directly create [] byte
	body := make([]byte, resp.ContentLength)
	defer func() {
		_ = resp.Body.Close()
	}()
	err = slices.ReadAllForByte(resp.Body, body)
	if err != nil {
		return nil, fmt.Errorf("ReadFile Err:[%+v]", err)
	}
	return &body, nil
}
