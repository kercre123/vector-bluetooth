package ble

import "strings"

// SendPin sends the pin and finishes the key exchange
func (v *VectorBLE) SendPin(pin string) error {
	// windows...
	if err := v.ble.SetPin(strings.TrimSpace(pin)); err != nil {
		return err
	}

	if err := v.ble.Send(v.state.getNonce()); err != nil {
		return err
	}
	v.ble.EnableEncryption()
	_, err := v.watch()
	return err
}
