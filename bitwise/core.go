package bitwise

import "math/rand"

func generateKey(seed string) []byte {
	// 使用 rand.NewSource 生成可重现的密钥
	source := rand.NewSource(int64(hashString(seed)))
	random := rand.New(source)

	keyLen := 32
	key := make([]byte, keyLen)

	for i := 0; i < keyLen; i++ {
		key[i] = byte(random.Intn(256))
	}

	return key
}

func hashString(s string) int {
	h := 0
	for _, c := range s {
		h = (h << 5) + int(c)
	}
	return h
}

func encrypt(data, key []byte) []byte {
	return apply(data, key)
}

func decrypt(data, key []byte) []byte {
	return apply(data, key)
}

func apply(data, key []byte) []byte {
	result := make([]byte, len(data))

	for i, b := range data {
		result[i] = b ^ key[i%len(key)] ^ key[(i+len(key)/2)%len(key)]
	}

	return result
}
