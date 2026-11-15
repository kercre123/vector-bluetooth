package conn

// Close shuts down the BLE connection
func (c *Connection) Close() error {
	c.connected.Disable()
	c.encrypted.Disable()
	c.established.Disable()
	if c.stopNotify != nil {
		c.stopNotify()
		c.stopNotify = nil
	}
	c.connection.Disconnect()
	return nil
}

// Reset clears all connection information
func (c *Connection) Reset() {
	c.connected.Disable()
	c.encrypted.Disable()
	c.established.Disable()
}
