package cluster

import (
	"modelcraft/pkg/crypto"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestCipher 初始化测试用的加密器
func setupTestCipher(t *testing.T) {
	// 使用32字节的测试密钥
	testKey := "12345678901234567890123456789012"
	err := crypto.InitDefaultAESCipher(testKey)
	require.NoError(t, err)
}

// TestNewByPlain_Success 测试使用明文创建密码对象成功
func TestNewByPlain_Success(t *testing.T) {
	setupTestCipher(t)

	plainPassword := "mySecretPassword123"
	password, err := NewByPlain(plainPassword)

	assert.NoError(t, err)
	assert.NotNil(t, password)
	assert.NotEmpty(t, password.encryptedValue)
	assert.NotNil(t, password.cipher)

	// 验证存储的是加密后的值，不是明文
	assert.NotEqual(t, plainPassword, password.encryptedValue)

	// 验证可以正确解密
	decrypted, err := password.GetPlainPassword()
	assert.NoError(t, err)
	assert.Equal(t, plainPassword, decrypted)
}

// TestNewByPlain_EmptyPassword 测试使用空明文创建密码对象失败
func TestNewByPlain_EmptyPassword(t *testing.T) {
	setupTestCipher(t)

	password, err := NewByPlain("")

	assert.Error(t, err)
	assert.Nil(t, password)
	assert.Contains(t, err.Error(), "plain password cannot be empty")
}

// TestNewByPlain_SpecialCharacters 测试使用包含特殊字符的明文创建密码对象
func TestNewByPlain_SpecialCharacters(t *testing.T) {
	setupTestCipher(t)

	specialPasswords := []string{
		"password!@#$%^&*()",
		"密码123",
		"pass word with spaces",
		"pass\nword\twith\rwhitespace",
		"🔐🔑🗝️",
	}

	for _, plainPassword := range specialPasswords {
		password, err := NewByPlain(plainPassword)

		assert.NoError(t, err, "Failed for password: %s", plainPassword)
		assert.NotNil(t, password)

		// 验证可以正确解密
		decrypted, err := password.GetPlainPassword()
		assert.NoError(t, err)
		assert.Equal(t, plainPassword, decrypted)
	}
}

// TestNewByPlain_LongPassword 测试使用长密码创建密码对象
func TestNewByPlain_LongPassword(t *testing.T) {
	setupTestCipher(t)

	// 创建一个很长的密码（1000个字符）
	longPassword := ""
	for i := 0; i < 100; i++ {
		longPassword += "0123456789"
	}

	password, err := NewByPlain(longPassword)

	assert.NoError(t, err)
	assert.NotNil(t, password)

	// 验证可以正确解密
	decrypted, err := password.GetPlainPassword()
	assert.NoError(t, err)
	assert.Equal(t, longPassword, decrypted)
}

// TestNewByEncrypt_Success 测试使用密文创建密码对象成功
func TestNewByEncrypt_Success(t *testing.T) {
	setupTestCipher(t)

	// 先创建一个加密的密码
	plainPassword := "testPassword123"
	cipher := crypto.GetDefaultAESCipher()
	encryptedValue, err := cipher.Encrypt(plainPassword)
	require.NoError(t, err)

	// 使用密文创建密码对象
	password, err := NewByEncrypt(encryptedValue)

	assert.NoError(t, err)
	assert.NotNil(t, password)
	assert.Equal(t, encryptedValue, password.encryptedValue)
	assert.NotNil(t, password.cipher)

	// 验证可以正确解密
	decrypted, err := password.GetPlainPassword()
	assert.NoError(t, err)
	assert.Equal(t, plainPassword, decrypted)
}

// TestNewByEncrypt_EmptyString 测试使用空字符串创建密码对象
func TestNewByEncrypt_EmptyString(t *testing.T) {
	setupTestCipher(t)

	password, err := NewByEncrypt("")

	assert.NoError(t, err)
	assert.NotNil(t, password)
	assert.Equal(t, "", password.encryptedValue)
}

// TestGetPlainPassword_Success 测试获取明文密码成功
func TestGetPlainPassword_Success(t *testing.T) {
	setupTestCipher(t)

	plainPassword := "myPassword123"
	password, err := NewByPlain(plainPassword)
	require.NoError(t, err)

	decrypted, err := password.GetPlainPassword()

	assert.NoError(t, err)
	assert.Equal(t, plainPassword, decrypted)
}

// TestGetPlainPassword_NilCipher 测试当cipher为nil时获取明文密码失败
func TestGetPlainPassword_NilCipher(t *testing.T) {
	password := &Password{
		encryptedValue: "someEncryptedValue",
		cipher:         nil, // cipher为nil
	}

	decrypted, err := password.GetPlainPassword()

	assert.Error(t, err)
	assert.Empty(t, decrypted)
	assert.Contains(t, err.Error(), "cipher is not available for decryption")
}

// TestGetPlainPassword_InvalidEncryptedValue 测试解密无效的密文失败
func TestGetPlainPassword_InvalidEncryptedValue(t *testing.T) {
	setupTestCipher(t)

	password := &Password{
		encryptedValue: "invalidEncryptedValue!!!",
		cipher:         crypto.GetDefaultAESCipher(),
	}

	decrypted, err := password.GetPlainPassword()

	assert.Error(t, err)
	assert.Empty(t, decrypted)
	assert.Contains(t, err.Error(), "failed to decrypt password")
}

// TestGetPlainPassword_EmptyEncryptedValue 测试解密空密文
func TestGetPlainPassword_EmptyEncryptedValue(t *testing.T) {
	setupTestCipher(t)

	password := &Password{
		encryptedValue: "",
		cipher:         crypto.GetDefaultAESCipher(),
	}

	decrypted, err := password.GetPlainPassword()

	assert.NoError(t, err)
	assert.Equal(t, "", decrypted)
}

// TestGetPassword_Success 测试获取存储的密码值
func TestGetPassword_Success(t *testing.T) {
	setupTestCipher(t)

	plainPassword := "testPassword"
	password, err := NewByPlain(plainPassword)
	require.NoError(t, err)

	encryptedValue := password.GetPassword()

	assert.NotEmpty(t, encryptedValue)
	assert.NotEqual(t, plainPassword, encryptedValue)
	assert.Equal(t, password.encryptedValue, encryptedValue)
}

// TestGetPassword_FromEncrypted 测试从密文创建的密码对象获取密码值
func TestGetPassword_FromEncrypted(t *testing.T) {
	setupTestCipher(t)

	encryptedValue := "someEncryptedValue"
	password, err := NewByEncrypt(encryptedValue)
	require.NoError(t, err)

	result := password.GetPassword()

	assert.Equal(t, encryptedValue, result)
}

// TestGetViewPasswd_AlwaysReturnsMask 测试获取显示密码总是返回掩码
func TestGetViewPasswd_AlwaysReturnsMask(t *testing.T) {
	setupTestCipher(t)

	testCases := []struct {
		name     string
		password string
	}{
		{"short password", "123"},
		{"normal password", "myPassword123"},
		{"long password", "thisIsAVeryLongPasswordWithManyCharacters123456789"},
		{"special chars", "!@#$%^&*()"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			password, err := NewByPlain(tc.password)
			require.NoError(t, err)

			viewPasswd := password.GetViewPasswd()

			assert.Equal(t, EncryptedByServerPlaceholder, viewPasswd)
		})
	}
}

// TestGetViewPasswd_FromEncrypted 测试从密文创建的密码对象获取显示密码
func TestGetViewPasswd_FromEncrypted(t *testing.T) {
	setupTestCipher(t)

	password, err := NewByEncrypt("anyEncryptedValue")
	require.NoError(t, err)

	viewPasswd := password.GetViewPasswd()

	assert.Equal(t, EncryptedByServerPlaceholder, viewPasswd)
}

// TestPassword_EncryptDecryptCycle 测试完整的加密解密周期
func TestPassword_EncryptDecryptCycle(t *testing.T) {
	setupTestCipher(t)

	originalPassword := "originalPassword123"

	// 1. 使用明文创建密码对象
	password1, err := NewByPlain(originalPassword)
	require.NoError(t, err)

	// 2. 获取加密后的值
	encryptedValue := password1.GetPassword()
	assert.NotEqual(t, originalPassword, encryptedValue)

	// 3. 使用加密值创建新的密码对象
	password2, err := NewByEncrypt(encryptedValue)
	require.NoError(t, err)

	// 4. 解密并验证
	decrypted, err := password2.GetPlainPassword()
	assert.NoError(t, err)
	assert.Equal(t, originalPassword, decrypted)
}

// TestPassword_MultipleInstances 测试多个密码对象互不影响
func TestPassword_MultipleInstances(t *testing.T) {
	setupTestCipher(t)

	password1, err := NewByPlain("password1")
	require.NoError(t, err)

	password2, err := NewByPlain("password2")
	require.NoError(t, err)

	password3, err := NewByPlain("password3")
	require.NoError(t, err)

	// 验证每个密码对象都是独立的
	decrypted1, err := password1.GetPlainPassword()
	assert.NoError(t, err)
	assert.Equal(t, "password1", decrypted1)

	decrypted2, err := password2.GetPlainPassword()
	assert.NoError(t, err)
	assert.Equal(t, "password2", decrypted2)

	decrypted3, err := password3.GetPlainPassword()
	assert.NoError(t, err)
	assert.Equal(t, "password3", decrypted3)

	// 验证加密值都不相同（因为使用了随机nonce）
	assert.NotEqual(t, password1.GetPassword(), password2.GetPassword())
	assert.NotEqual(t, password2.GetPassword(), password3.GetPassword())
	assert.NotEqual(t, password1.GetPassword(), password3.GetPassword())
}

// TestPassword_SamePlainDifferentEncrypted 测试相同明文产生不同密文
func TestPassword_SamePlainDifferentEncrypted(t *testing.T) {
	setupTestCipher(t)

	plainPassword := "samePassword"

	// 创建两个使用相同明文的密码对象
	password1, err := NewByPlain(plainPassword)
	require.NoError(t, err)

	password2, err := NewByPlain(plainPassword)
	require.NoError(t, err)

	// 由于使用了随机nonce，相同的明文应该产生不同的密文
	assert.NotEqual(t, password1.GetPassword(), password2.GetPassword())

	// 但解密后应该得到相同的明文
	decrypted1, err := password1.GetPlainPassword()
	assert.NoError(t, err)

	decrypted2, err := password2.GetPlainPassword()
	assert.NoError(t, err)

	assert.Equal(t, plainPassword, decrypted1)
	assert.Equal(t, plainPassword, decrypted2)
}

// TestPassword_Immutability 测试密码对象的不可变性
func TestPassword_Immutability(t *testing.T) {
	setupTestCipher(t)

	plainPassword := "immutablePassword"
	password, err := NewByPlain(plainPassword)
	require.NoError(t, err)

	// 获取加密值
	encryptedValue1 := password.GetPassword()
	encryptedValue2 := password.GetPassword()

	// 多次获取应该返回相同的值
	assert.Equal(t, encryptedValue1, encryptedValue2)

	// 解密多次应该返回相同的明文
	decrypted1, err := password.GetPlainPassword()
	assert.NoError(t, err)

	decrypted2, err := password.GetPlainPassword()
	assert.NoError(t, err)

	assert.Equal(t, plainPassword, decrypted1)
	assert.Equal(t, plainPassword, decrypted2)
}
