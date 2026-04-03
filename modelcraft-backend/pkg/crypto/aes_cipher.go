package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// AESCipher AES加密器
type AESCipher struct {
	key []byte
}

var defaultAESCipher *AESCipher

// InitDefaultAESCipher 初始化默认的AES加密器
// key 必须是32字节长度用于AES-256加密
func InitDefaultAESCipher(key string) error {
	cipher, err := NewAESCipher(key)
	if err != nil {
		return err
	}
	defaultAESCipher = cipher
	return nil
}

// GetDefaultAESCipher 获取默认的AES加密器实例
// 如果未初始化会触发panic
func GetDefaultAESCipher() *AESCipher {
	if defaultAESCipher == nil {
		panic("AES cipher not initialized")
	}
	return defaultAESCipher
}

// NewAESCipher 创建新的AES加密器
// key 必须是32字节长度用于AES-256加密
func NewAESCipher(key string) (*AESCipher, error) {
	keyBytes := []byte(key)
	if len(keyBytes) != 32 {
		return nil, errors.New("key must be 32 bytes for AES-256")
	}
	return &AESCipher{key: keyBytes}, nil
}

// Encrypt 加密明文字符串
func (a *AESCipher) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	// 创建AES cipher
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return "", err
	}

	// 创建GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 生成随机nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// 加密数据
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// 返回base64编码的密文
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密密文字符串
func (a *AESCipher) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	// 解码base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	// 创建AES cipher
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return "", err
	}

	// 创建GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 检查数据长度
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	// 分离nonce和密文
	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]

	// 解密数据
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// GenerateKey 生成32字节的随机密钥
func GenerateKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}
