package blecrypto

import (
	"golang.org/x/crypto/chacha20poly1305"
)

// Encrypt encrypts an incoming message
func (b *BLECrypto) Encrypt(msg []byte) ([]byte, error) {
	aead, err := chacha20poly1305.NewX(b.encrypt[:])
	if err != nil {
		return nil, err
	}

	c := aead.Seal(nil, b.encryptionNonce[:], msg, nil)
	b.nextEncryptNonce()
	return c, nil
}
