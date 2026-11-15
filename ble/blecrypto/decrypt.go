package blecrypto

import (
	"errors"

	"golang.org/x/crypto/chacha20poly1305"
)

const (
	aBytes int = chacha20poly1305.Overhead // Size of an authentication tag in bytes
)

// DecryptMessage decrypts an incoming message
func (b *BLECrypto) DecryptMessage(msg []byte) ([]byte, error) {
	if len(msg) < aBytes {
		return nil, errors.New("message too short")
	}

	aead, err := chacha20poly1305.NewX(b.decrypt[:])
	if err != nil {
		return nil, err
	}

	m, err := aead.Open(nil, b.decryptionNonce[:], msg, nil)
	b.nextDecryptNonce()
	if err != nil {
		return nil, errors.New("verification failed")
	}

	return m, nil
}
