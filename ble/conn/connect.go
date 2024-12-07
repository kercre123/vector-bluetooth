package conn

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/pkg/errors"
	"tinygo.org/x/bluetooth"
)

const (
	duration   = 2
	readUUID   = "7d2a4bda-d29b-4152-b725-2491478c5cd7"
	writeUUID  = "30619f2d-0f54-41bd-a65a-7588d8c85b45"
	retryCount = 3
	offset     = 1
)

var blebuf bleBuffer

// Connect connects to a specific device
func (c *Connection) Connect(id int) error {
	if err := c.bleConnect(id); err != nil {
		return err
	}

	if err := retry(retryCount, time.Second, c.discoverProfile); err != nil {
		return err
	}

	if err := retry(retryCount, time.Second, c.findReader); err != nil {
		return err
	}

	if err := retry(retryCount, time.Second, c.findWriter); err != nil {
		return err
	}

	errCh := make(chan error)
	go c.subscribe(errCh)
	err := <-errCh
	if err != nil {
		return err
	}
	c.established.Enable()

	go c.handleIncoming()
	return nil
}

// bleConnect handles establishing the actual connection
func (c *Connection) bleConnect(id int) error {
	_, cancel := context.WithTimeout(context.Background(), scanDuration*duration)
	defer cancel()

	addr := c.scanresults.getresult(id)
	if addr == nil {
		return errors.New("no device found with that ID")
	}

	dev, err := bluetooth.DefaultAdapter.Connect(*addr, c.connParams)
	if err != nil {
		return err
	}

	c.connection = dev
	return nil
}

func ParseUUID(s string) (bluetooth.UUID, error) {
	uuid, err := bluetooth.ParseUUID(s)
	if err != nil {
		return bluetooth.UUID{}, err
	}
	return uuid, nil
}

// discoverProfile finds the device services and characteristics
func (c *Connection) discoverProfile() error {
	services, err := c.connection.DiscoverServices(nil)
	if err != nil {
		return errors.Wrap(err, "can't discover services")
	}
	c.services = services
	return nil
}

// findWriter configures the writer characteristic
func (c *Connection) findWriter() error {
	wUUID, err := ParseUUID(writeUUID)
	if err != nil {
		return errors.Wrap(err, "invalid writer UUID")
	}

	for _, svc := range c.services {
		chars, err := svc.DiscoverCharacteristics(nil)
		if err != nil {
			continue
		}
		for i := range chars {
			if chars[i].UUID().String() == wUUID.String() {
				c.writer = &chars[i]
				return nil
			}
		}
	}
	return errors.New("cannot find write channel")
}

// findReader configures the reader characteristic
func (c *Connection) findReader() error {
	rUUID, err := ParseUUID(readUUID)
	if err != nil {
		return errors.Wrap(err, "invalid reader UUID")
	}

	for _, svc := range c.services {
		chars, err := svc.DiscoverCharacteristics(nil)
		if err != nil {
			continue
		}
		for i := range chars {
			if chars[i].UUID().String() == rUUID.String() {
				c.reader = &chars[i]
				return nil
			}
		}
	}
	return errors.New("cannot find read channel")
}

// subscribe pipes incoming data to a reader chan
func (c *Connection) subscribe(errChan chan error) {
	if c.writer == nil {
		errChan <- errors.New("writer characteristic not found")
		return
	}

	// writer/reader are deceiving!

	err := c.writer.EnableNotifications(func(buf []byte) {
		c.incoming <- buf
	})
	errChan <- err
}

func (c *Connection) handleIncoming() {
	for incoming := range c.incoming {
		if incoming == nil {
			continue
		}
		b := blebuf.receiveRawBuffer(incoming)
		if b == nil {
			continue
		}
		switch {
		case !c.connected.Enabled():
			c.handleConnectionRequest(incoming)
		case !c.encrypted.Enabled() && c.connected.Enabled():
			c.out <- b
		case c.encrypted.Enabled() && c.connected.Enabled():
			buf, err := c.crypto.DecryptMessage(b)
			if err != nil {
				fmt.Println("ERROR", err)
			}
			c.out <- buf
		default:
			c.established.Disable()
			c.encrypted.Disable()
		}
	}
}

func (c *Connection) handleConnectionRequest(buffer []byte) {
	if err := c.rawMessage(buffer); err != nil {
		return
	}
	c.connected.Enable()
	c.version = int(buffer[2])
}

func retry(attempts int, sleep time.Duration, f func() error) error {
	if err := f(); err != nil {
		attempts--
		if attempts > 0 {
			jitter := time.Duration(rand.Int63n(int64(sleep)))
			sleep += jitter / offset
			time.Sleep(sleep)
			return retry(attempts, offset*sleep, f)
		}
		return err
	}
	return nil
}
