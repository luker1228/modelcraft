package cluster

import (
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/crypto"
)

// EncryptedByServerPlaceholder is the sentinel value returned by GetViewPasswd.
// When the client sends this value back in an update request, the server treats it
// as "password unchanged" and keeps the existing encrypted password in the database.
const EncryptedByServerPlaceholder = "<encrypted_by_server>"

// Password 封装密码值，区分明文和加密状态
type Password struct {
	encryptedValue string            // 存储的值（明文或加密）
	cipher         *crypto.AESCipher // AES加密器实例，用于解密
}

// NewByEncrypt 入参为密文，直接构造
func NewByEncrypt(encryptedPasswd string) (*Password, error) {
	cipher := crypto.GetDefaultAESCipher()
	return &Password{
		encryptedValue: encryptedPasswd,
		cipher:         cipher,
	}, nil
}

// NewByPlain 入参为明文，需要加密后存储
func NewByPlain(plainPassword string) (*Password, error) {
	if plainPassword == "" {
		return nil, bizerrors.Errorf("plain password cannot be empty")
	}

	cipher := crypto.GetDefaultAESCipher()
	encrypted, err := cipher.Encrypt(plainPassword)
	if err != nil {
		return nil, bizerrors.Errorf("failed to encrypt password: %w", err)
	}

	return &Password{
		encryptedValue: encrypted,
		cipher:         cipher,
	}, nil
}

// GetPlainPassword 获取明文密码
func (p *Password) GetPlainPassword() (string, error) {
	if p.cipher == nil {
		return "", bizerrors.Errorf("cipher is not available for decryption")
	}

	// 解密密码
	plain, err := p.cipher.Decrypt(p.encryptedValue)
	if err != nil {
		return "", bizerrors.Errorf("failed to decrypt password: %w", err)
	}

	return plain, nil
}

// GetPassword 获取存储的密码值
func (p *Password) GetPassword() string {
	return p.encryptedValue
}

// GetViewPasswd returns the sentinel placeholder for display purposes.
// Clients must send this value back unchanged to indicate the password has not been modified.
func (p *Password) GetViewPasswd() string {
	return EncryptedByServerPlaceholder
}
