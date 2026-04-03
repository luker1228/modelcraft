package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodePasswd(t *testing.T) {
	cipher, err := NewAESCipher("12345678901234567890123456789012")
	if err != nil {
		t.Fatal(err)
	}
	encryptedcode, err := cipher.Encrypt("password")
	if err != nil {
		t.Fatal(err)
	}
	plainpasswd, err := cipher.Decrypt(encryptedcode)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "password", plainpasswd)
}
