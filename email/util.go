package email

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// ReadAllForByte rewrite io.ReadAll
func readAllForByte(r io.Reader, b []byte) error {
	b = b[:0]
	for {
		n, err := r.Read(b[len(b):cap(b)])
		b = b[:len(b)+n]
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return err
		}
	}
}

//getFileContent 获取文件内容
/**
通过判断 url 是否包含 http:// || https:// 来判断是否使用 http下载
*/
func getFileContent(url string) ([]byte, error) {
	if strings.Contains(url, "http://") || strings.Contains(url, "https://") {
		b, err := downloadFile(url)
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

// getAttachmentName 获取发送附件文件名称
func getAttachmentName(path string) string {
	names := strings.Split(path, "|") //判断是否有需要自定义附件名称 xxxxx|fileName
	if len(names) == 2 {
		return names[1]
	}
	names = strings.Split(path, "/") //无自定义fileName 默认使用路径/后最后部分
	return names[len(names)-1]
}

// downloadFile 下载文件
func downloadFile(url string) (*[]byte, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK || resp.Body == nil {
		return nil, err
	}
	//body := make([]byte, 0, resp.ContentLength)
	body := make([]byte, resp.ContentLength)
	defer func() {
		_ = resp.Body.Close()
	}()
	//body, err = ReadAll(resp.Body, body)
	err = readAllForByte(resp.Body, body)
	if err != nil {
		return nil, fmt.Errorf("ReadFile Err:[%+v]", err)
	}
	return &body, nil
}
