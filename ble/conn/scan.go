package conn

import (
	"context"
	"time"

	"tinygo.org/x/bluetooth"
)

const (
	scanDuration = 3 * time.Second
)

// ScanResponse is a list of devices the BLE adaptor has found
type ScanResponse struct {
	Devices []*Device
}

// Device is a single device entity
type Device struct {
	ID      int
	Name    string
	Address string
}

// Scan looks for BLE devices matching the vector requirements
func (c *Connection) Scan() (*ScanResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), scanDuration)
	defer cancel()

	done := make(chan struct{})
	var scanErr error
	go func() {
		scanErr = c.device.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
			c.scan(result)
		})
		close(done)
	}()

	<-ctx.Done()

	_ = c.device.StopScan()
	<-done

	if scanErr != nil {
		return nil, scanErr
	}

	d := []*Device{}
	for k, v := range c.scanresults.getresults() {
		td := Device{
			ID:      k,
			Name:    v.name,
			Address: v.addr.String(),
		}
		d = append(d, &td)
	}

	resp := ScanResponse{
		Devices: d,
	}

	return &resp, nil
}
