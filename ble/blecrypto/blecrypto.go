package blecrypto

import (
	"crypto/rand"

	"golang.org/x/crypto/curve25519"
)

// BLECrypto is a container for all of the BLE crypto
type BLECrypto struct {
	secretKey       [32]byte
	publicKey       [32]byte
	remotePublicKey []byte
	encryptionNonce [24]uint8
	decryptionNonce [24]uint8
	decrypt         [32]byte
	encrypt         [32]byte
}

// New returns a BLECrypto with a populated keyset
func New() *BLECrypto {
	b := &BLECrypto{}
	if _, err := rand.Read(b.secretKey[:]); err != nil {
		panic(err)
	}
	pk, err := curve25519.X25519(b.secretKey[:], curve25519.Basepoint)
	if err != nil {
		panic(err)
	}
	copy(b.publicKey[:], pk)
	return b
}
