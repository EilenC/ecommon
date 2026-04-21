package bitwise

import (
	"encoding/gob"
	"errors"
	"github.com/EilenC/ecommon"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Encrypted struct {
	Data []byte
}

var ext = ".bitwise"

func SetExt(e string) {
	ext = e
}

// SaveFile 保存文件
func SaveFile(encrypted []byte, outPutFile string) (string, error) {
	outPutFile = outPutFile + ext
	file, err := ecommon.CreateFile(outPutFile)
	if err != nil {
		return "", err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	return filepath.Base(outPutFile), encodeEncrypted(file, encrypted)
}

func encodeEncrypted(w io.Writer, encrypted []byte) error {
	if err := gob.NewEncoder(w).Encode(Encrypted{Data: encrypted}); err != nil {
		return errors.New("gob encode err " + err.Error())
	}
	return nil
}

// GetRealFileName 获取真实文件名称(去掉路径前缀与加密后的ext)
func GetRealFileName(filePath string) string {
	// 使用 filepath.Base 获取文件路径中的最后一个元素
	fileName := filepath.Base(filePath)

	// 使用 strings.TrimSuffix 去掉文件名的后缀
	return strings.TrimSuffix(fileName, ext)
}
