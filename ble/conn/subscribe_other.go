//go:build !linux

package conn

import "github.com/pkg/errors"

func (c *Connection) subscribe(errChan chan error) {
	if c.writer == nil {
		errChan <- errors.New("writer characteristic not found")
		return
	}

	err := c.writer.EnableNotifications(func(buf []byte) {
		b := make([]byte, len(buf))
		copy(b, buf)
		c.incoming <- b
	})
	errChan <- err
}
