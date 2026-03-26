package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCryptoService_EmptyCases(t *testing.T) {
	key := "12345678901234567890123456789012"
	service, _ := New(key)

	res, err := service.Encrypt([]byte{})
	assert.NoError(t, err)
	assert.Empty(t, res)

	res, err = service.Decrypt([]byte{})
	assert.NoError(t, err)
	assert.Nil(t, res)

	res, err = service.Decrypt([]byte("short"))
	assert.Error(t, err)
	assert.Nil(t, res)
}
