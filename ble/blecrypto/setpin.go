package blecrypto

import (
	"errors"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/curve25519"
)

const (
	keysize = 32
)

// SetPin configures the hashes based on the pin entered
func (b *BLECrypto) SetPin(pin string) error {
	if b.remotePublicKey == nil {
		return errors.New("remote public key is not set")
	}

	rx, tx, err := b.clientSessionKeys()
	if err != nil {
		return err
	}
	b.decrypt, err = genHash(rx, []byte(pin))
	if err != nil {
		return err
	}
	b.encrypt, err = genHash(tx, []byte(pin))
	return err
}

func (b *BLECrypto) clientSessionKeys() (rx, tx []byte, err error) {
	q, err := curve25519.X25519(b.secretKey[:], b.remotePublicKey)
	if err != nil {
		return nil, nil, err
	}
	in := make([]byte, 0, 3*keysize)
	in = append(in, q...)
	in = append(in, b.publicKey[:]...)
	in = append(in, b.remotePublicKey...)
	k := blake2b.Sum512(in)
	return k[:keysize], k[keysize:], nil
}

func genHash(key, pin []byte) ([32]byte, error) {
	rv := [32]byte{}
	h, err := blake2b.New(keysize, pin)
	if err != nil {
		return rv, err
	}
	h.Write(key)
	copy(rv[:], h.Sum(nil))
	return rv, nil
}
