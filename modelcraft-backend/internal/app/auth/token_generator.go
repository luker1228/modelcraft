package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// GenerateRefreshToken 生成 32 字节 CSPRNG → 64 位 hex 字符串。
// 返回明文 token 和其 SHA256 hash（hash 存 DB，明文返回客户端）。
func GenerateRefreshToken() (plaintext, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}
	plaintext = hex.EncodeToString(b)
	sum := sha256.Sum256([]byte(plaintext))
	hash = hex.EncodeToString(sum[:])
	return plaintext, hash, nil
}

// base62Chars Base62 字符集（0-9, A-Z, a-z）
const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// GenerateAPIKey 生成 API Key：mc_ 前缀 + 40 位 Base62（CSPRNG，约 238 bit 熵）。
// 返回完整明文 key 和其 SHA256 hash。
func GenerateAPIKey() (plaintext, hash string, err error) {
	const length = 40
	b := make([]byte, length)
	if _, err = rand.Read(b); err != nil {
		return "", "", fmt.Errorf("generate api key: %w", err)
	}
	result := make([]byte, length)
	for i, v := range b {
		result[i] = base62Chars[int(v)%62]
	}
	plaintext = "mc_" + string(result)
	sum := sha256.Sum256([]byte(plaintext))
	hash = hex.EncodeToString(sum[:])
	return plaintext, hash, nil
}

// HashToken 对任意 token 字符串计算 SHA256 hash。
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
