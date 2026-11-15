package blecrypto

import (
	"github.com/kercre123/vector-bluetooth/rts"
)

// SetRemotePublicKey populates the remote public key
func (b *BLECrypto) SetRemotePublicKey(msg *rts.RtsConnRequest) error {
	b.remotePublicKey = append([]byte{}, msg.PublicKey[:]...)
	return nil
}

// GetRemotePublicKey returns the public key
func (b *BLECrypto) GetRemotePublicKey() [32]uint8 {
	return b.publicKey
}
