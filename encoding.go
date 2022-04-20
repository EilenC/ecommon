package go_common

import (
	"bytes"
	"io/ioutil"

	"golang.org/x/net/html/charset"
)

//ConvertToUTF8 指定编码转换为UTF8
//或使用 https://github.com/djimenez/iconv-go
func ConvertToUTF8(body []byte, origEncoding string) string {
	byteReader := bytes.NewReader(body)
	reader, _ := charset.NewReaderLabel(origEncoding, byteReader)
	body, _ = ioutil.ReadAll(reader)
	return string(body)
}
