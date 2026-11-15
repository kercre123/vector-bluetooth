package blecrypto

import (
	"github.com/kercre123/vector-bluetooth/rts"
)

// SetNonces populates the nonces
func (b *BLECrypto) SetNonces(msg *rts.RtsNonceMessage) error {
	b.encryptionNonce = msg.ToRobotNonce
	b.decryptionNonce = msg.ToDeviceNonce
	return nil
}

func (b *BLECrypto) nextEncryptNonce() {
	increment(b.encryptionNonce[:])
}

func (b *BLECrypto) nextDecryptNonce() {
	increment(b.decryptionNonce[:])
}

func increment(n []byte) {
	c := uint16(1)
	for i := range n {
		c += uint16(n[i])
		n[i] = byte(c)
		c >>= 8
	}
}
