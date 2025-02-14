package secret

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrapSecret(t *testing.T) {
	secret := "mysecret"
	masterKey := "myverystrongpasswordo32bitlength"

	wrappedSecret, err := WrapSecret(secret, masterKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, wrappedSecret)
}

func TestUnwrapSecret(t *testing.T) {
	secret := "mysecret"
	masterKey := "myverystrongpasswordo32bitlength"

	wrappedSecret, err := WrapSecret(secret, masterKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, wrappedSecret)

	unwrappedSecret, err := UnwrapSecret(wrappedSecret, masterKey)
	assert.NoError(t, err)
	assert.Equal(t, UnwrappedSecret(secret), unwrappedSecret)
}
