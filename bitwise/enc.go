package bitwise

import (
	"errors"
	"os"
)

// EncryptFile 指定文件进行加密
// filePath 指定要加密的文件路径
// seed 类似key一样 作用与随机生成 key (加密解密要保持一致)
// outPutPath 指定加密后的gob文件存储地址(会自动添加上ext中设置的值)
func EncryptFile(filePath, seed, prePath string) (string, error) {
	if len(seed) == 0 {
		return "", errors.New("seed value is invalid")
	}
	fullFilePath := prePath + string(os.PathSeparator) + filePath
	data, err := os.ReadFile(fullFilePath)
	if err != nil {
		return "", err
	}
	b, err := Encrypt(data, seed)
	if err != nil {
		return "", err
	}
	fName, err := SaveFile(b, fullFilePath)
	if err != nil {
		return "", err
	}
	return fName, nil
}

func Encrypt(b []byte, seed string) ([]byte, error) {
	if len(seed) == 0 {
		return nil, errors.New("seed value is invalid")
	}
	encK := generateKey(seed)
	data := make([]byte, len(b))
	copy(data, b)
	//bitwise
	encrypted := make([]byte, len(data))
	encrypted = encrypt(data, encK)

	return encrypted, nil
}
