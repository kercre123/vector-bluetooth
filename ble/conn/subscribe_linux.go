//go:build linux

package conn

import (
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/pkg/errors"
)

const (
	gattCharIface = "org.bluez.GattCharacteristic1"
	propsChanged  = "org.freedesktop.DBus.Properties.PropertiesChanged"
)

// godbus function but it doesn't reorder data
func (c *Connection) subscribe(errChan chan error) {
	if c.writer == nil {
		errChan <- errors.New("writer characteristic not found")
		return
	}

	bus, err := dbus.SystemBus()
	if err != nil {
		errChan <- errors.Wrap(err, "can't get system bus")
		return
	}

	path, err := findCharPath(bus, c.connection.Address.String(), writeUUID)
	if err != nil {
		errChan <- err
		return
	}

	sigs := make(chan *dbus.Signal, 4096)
	bus.Signal(sigs)
	match := []dbus.MatchOption{
		dbus.WithMatchObjectPath(path),
		dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
		dbus.WithMatchMember("PropertiesChanged"),
	}
	if err := bus.AddMatchSignal(match...); err != nil {
		bus.RemoveSignal(sigs)
		errChan <- errors.Wrap(err, "can't add match signal")
		return
	}

	obj := bus.Object("org.bluez", path)
	if err := obj.Call(gattCharIface+".StartNotify", 0).Err; err != nil {
		bus.RemoveMatchSignal(match...)
		bus.RemoveSignal(sigs)
		errChan <- errors.Wrap(err, "can't start notify")
		return
	}

	c.stopNotify = func() {
		obj.Call(gattCharIface+".StopNotify", 0)
		bus.RemoveMatchSignal(match...)
		bus.RemoveSignal(sigs)
		close(sigs)
	}

	go func() {
		for sig := range sigs {
			if sig.Path != path || sig.Name != propsChanged || len(sig.Body) < 2 {
				continue
			}
			if iface, _ := sig.Body[0].(string); iface != gattCharIface {
				continue
			}
			changes, ok := sig.Body[1].(map[string]dbus.Variant)
			if !ok {
				continue
			}
			val, ok := changes["Value"]
			if !ok {
				continue
			}
			if b, ok := val.Value().([]byte); ok {
				c.incoming <- b
			}
		}
	}()

	errChan <- nil
}

func findCharPath(bus *dbus.Conn, addr, uuid string) (dbus.ObjectPath, error) {
	var objs map[dbus.ObjectPath]map[string]map[string]dbus.Variant
	err := bus.Object("org.bluez", "/").Call("org.freedesktop.DBus.ObjectManager.GetManagedObjects", 0).Store(&objs)
	if err != nil {
		return "", errors.Wrap(err, "can't list bluez objects")
	}

	dev := "/dev_" + strings.ReplaceAll(strings.ToUpper(addr), ":", "_") + "/"
	for p, ifaces := range objs {
		props, ok := ifaces[gattCharIface]
		if !ok || !strings.Contains(string(p), dev) {
			continue
		}
		u, _ := props["UUID"].Value().(string)
		if strings.EqualFold(u, uuid) {
			return p, nil
		}
	}
	return "", errors.New("can't find notify characteristic on dbus")
}
