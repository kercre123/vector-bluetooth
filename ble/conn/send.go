package conn

const (
	maxSize = 20
)

// Send sends a message via BLE
func (c *Connection) Send(buf []byte) error {
	if c.encrypted.Enabled() {
		var err error
		buf, err = c.crypto.Encrypt(buf)
		if err != nil {
			return err
		}
	}

	size := len(buf)
	if size < maxSize {
		return c.messageWithHeader(msgSolo, buf, size)
	}

	remaining := len(buf)

	for remaining > 0 {
		offset := size - remaining

		switch {
		case remaining == size:
			// first chunk
			msgSize := maxSize - 1
			if err := c.messageWithHeader(msgStart, buf[offset:offset+msgSize], msgSize); err != nil {
				return err
			}
			remaining -= msgSize

		case remaining < maxSize:
			// last chunk
			if err := c.messageWithHeader(msgEnd, buf[offset:], remaining); err != nil {
				return err
			}
			remaining = 0

		default:
			// middle chunk
			msgSize := maxSize - 1
			if err := c.messageWithHeader(msgContinue, buf[offset:offset+msgSize], msgSize); err != nil {
				return err
			}
			remaining -= msgSize
		}
	}
	return nil
}

func (c *Connection) messageWithHeader(multipart byte, buffer []byte, size int) error {
	var msg []byte
	msg = append(msg, getHeaderByte(multipart, size))
	msg = append(msg, buffer...)

	return c.rawMessage(msg)
}

func (c *Connection) rawMessage(buffer []byte) error {
	if c.reader == nil {
		return nil
	}
	_, err := c.reader.WriteWithoutResponse(buffer)
	return err
}

func getHeaderByte(multipart byte, size int) byte {
	return byte(((int(multipart) << 6) | (size & 0x3F)))
}
