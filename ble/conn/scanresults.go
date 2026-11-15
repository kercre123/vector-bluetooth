package conn

import (
	"strings"
	"sync"

	"tinygo.org/x/bluetooth"
)

const vectorservice = "fee3"

type result struct {
	name string
	addr *bluetooth.Address
}

type scan struct {
	results      sync.Map
	mutexCounter *mCounter
}

func newScan() *scan {
	m := scan{
		results:      sync.Map{},
		mutexCounter: newMutexCounter(),
	}
	return &m
}

// windows doesn't give us service uuids, so match the name too
func isVector(d bluetooth.ScanResult) bool {
	if d.HasServiceUUID(mustParseShortUUID(vectorservice)) {
		return true
	}
	return isVectorName(d.LocalName())
}

// "Vector U1R7" yes, "Vector Cube" no
func isVectorName(name string) bool {
	suffix, ok := strings.CutPrefix(name, "Vector ")
	if !ok || len(suffix) != 4 {
		return false
	}
	for i := 0; i < len(suffix); i++ {
		b := suffix[i]
		if !(b >= 'A' && b <= 'Z') && !(b >= '0' && b <= '9') {
			return false
		}
	}
	return true
}

func (c *Connection) scan(d bluetooth.ScanResult) {
	if isVector(d) {
		// dedup
		r := c.scanresults.getresults()
		for _, v := range r {
			if d.Address.String() == v.addr.String() {
				return
			}
		}
		c.scanresults.results.Store(c.scanresults.mutexCounter.getCount(), result{
			name: d.LocalName(),
			addr: &d.Address,
		})
	}
}

func (m *scan) getresults() map[int]result {
	tm := map[int]result{}
	m.results.Range(
		func(key, value interface{}) bool {
			r := value.(result)
			tm[key.(int)] = r
			return true
		},
	)
	return tm
}

func (m *scan) getresult(id int) *bluetooth.Address {
	v, ok := m.results.Load(id)
	if !ok {
		return nil
	}
	r := v.(result)
	return r.addr
}

type mCounter struct {
	count int
	m     sync.Mutex
}

func newMutexCounter() *mCounter {
	return &mCounter{
		count: 1,
		m:     sync.Mutex{},
	}
}

func (m *mCounter) getCount() int {
	m.m.Lock()
	r := m.count
	m.count++
	m.m.Unlock()
	return r
}

func mustParseShortUUID(short string) bluetooth.UUID {
	var val uint16
	for i := 0; i < len(short); i++ {
		val <<= 4
		b := short[i]
		switch {
		case b >= '0' && b <= '9':
			val |= uint16(b - '0')
		case b >= 'a' && b <= 'f':
			val |= uint16(b-'a') + 10
		case b >= 'A' && b <= 'F':
			val |= uint16(b-'A') + 10
		}
	}
	return bluetooth.New16BitUUID(val)
}
