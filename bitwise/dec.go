package bitwise

import (
	"encoding/gob"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// DecryptFile 指定文件进行解密
// filePath 指定要解密的文件路径(包含ext的后缀名)
// seed 类似key一样 作用与随机生成 key (加密解密要保持一致)
// outPutPath 指定解密后的源文件存储地址 (会自动去除ext补充的后缀名)
//func DecryptFile(filePath, seed, outPutPath string) error {
//	dec, err := Decrypt(filePath, seed)
//	if err != nil {
//		return err
//	}
//	return os.WriteFile(strings.ReplaceAll(outPutPath, ext, ""), dec, os.ModePerm)
//}

// Decrypt 解密 []byte 数据
func Decrypt(encB []byte, seed string) ([]byte, error) {
	if len(seed) == 0 {
		return nil, errors.New("seed value is invalid")
	}
	encK := generateKey(seed)
	data := make([]byte, len(encB))
	copy(data, encB)
	//bitwise
	decrypted := make([]byte, len(data))
	decrypted = decrypt(data, encK)

	return decrypted, nil
}

// DecryptFile 指定文件进行解密
func DecryptFile(filePath, seed, filePrefix string) (string, error) {
	// 解密文件数据
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	gobD := gob.NewDecoder(file)

	var decodedData Encrypted
	err = gobD.Decode(&decodedData)
	if err != nil {
		return "", errors.New("gob decode err " + err.Error())
	}
	if len(decodedData.Data) == 0 {
		return "", errors.New("file is empty")
	}
	b, err := Decrypt(decodedData.Data, seed)
	if err != nil {
		return "", err
	}
	if len(b) == 0 {
		return "", errors.New("decrypted file is empty")
	}
	suffixExtName := filePrefix + GetRealFileName(filePath)
	err = os.WriteFile(strings.Join([]string{filepath.Dir(filePath), suffixExtName}, string(os.PathSeparator)), b, os.ModePerm)
	if err != nil {
		return "", err
	}
	return suffixExtName, nil
}
