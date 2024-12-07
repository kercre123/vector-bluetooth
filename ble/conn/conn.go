package conn

import (
	"math/rand"
	"time"

	"github.com/digital-dream-labs/vector-bluetooth/ble/blecrypto"
	"github.com/pkg/errors"
	"tinygo.org/x/bluetooth"
)

var DefaultAdapter = &bluetooth.Adapter{}

// Connection is the configuration struct for ble connections
type Connection struct {
	device      *bluetooth.Adapter
	scanresults *scan
	connection  bluetooth.Device
	services    []bluetooth.DeviceService
	reader      *bluetooth.DeviceCharacteristic
	writer      *bluetooth.DeviceCharacteristic
	incoming    chan []byte
	out         chan []byte
	crypto      *blecrypto.BLECrypto
	version     int
	established lockState
	connected   lockState
	encrypted   lockState
	connParams  bluetooth.ConnectionParams
}

// New returns a connection, or an error on failure
func New(output chan []byte) (*Connection, error) {
	rand.Seed(time.Now().UnixNano())

	// Initialize the default adapter
	DefaultAdapter = bluetooth.DefaultAdapter
	if err := DefaultAdapter.Enable(); err != nil {
		return nil, errors.Wrap(err, "can't enable default adapter")
	}

	c := Connection{
		device:      DefaultAdapter,
		scanresults: newScan(),
		incoming:    make(chan []byte),
		out:         output,
		crypto:      blecrypto.New(),
		connParams: bluetooth.ConnectionParams{
			ConnectionTimeout: bluetooth.NewDuration(time.Second * 3), // default
			MinInterval:       0,                                      // default
			MaxInterval:       0,                                      // default
		},
	}

	return &c, nil
}

// EnableEncryption sets the encryption bit to automatically de/encrypt
func (c *Connection) EnableEncryption() {
	c.encrypted.Enable()
}

// Connected lets external packages know if the initial connection attempt has happened
func (c *Connection) Connected() bool {
	return c.connected.Enabled()
}
