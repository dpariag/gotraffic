package network

import (
	"git.svc.rocks/dpariag/gotraffic/stats"
	"github.com/google/gopacket"
	"time"
)

// Implements ioHandle interface for unit testing
type testIOHandle struct {
	handle chan []byte
	stats  stats.Directional
}

func (t *testIOHandle) ReadPacketData() (data []byte, ci gopacket.CaptureInfo, err error) {
	data = <-t.handle
	ci.Timestamp = time.Now()
	ci.CaptureLength = len(data)
	ci.Length = len(data)
	t.stats.Rx.Packets++
	t.stats.Rx.Bytes += uint64(len(data))
	return data, ci, nil
}

func (t *testIOHandle) WritePacketData(data []byte) (err error) {
	t.handle <- data
	t.stats.Tx.Packets++
	t.stats.Tx.Bytes += uint64(len(data))
	return nil
}

func (t *testIOHandle) Stats() stats.Directional {
	return t.stats
}

func newTestIOHandle() *testIOHandle {
	return &testIOHandle{handle: make(chan []byte, 2048)}
}
