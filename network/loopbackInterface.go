// A conceptual loopback interface, intended primarily for unit testing
// Packets written to this interface are immediately returned to the interface
// Registered listeners can be notified immediately
// Requires no underlying NIC

package network

import (
	"time"
	"code.google.com/p/gopacket"
)

//TODO: Embed interface stats, rather than copying them from regular interface
type LoopbackInterface struct {
	rxChannels map[uint64] chan gopacket.Packet
	rxPkts	uint64
	rxBytes	uint64
	txPkts	uint64
	txBytes	uint64
}

func NewLoopback() *LoopbackInterface {
	i := LoopbackInterface{}
	i.rxChannels = make(map[uint64]chan gopacket.Packet)
	return &i
}

// Register a (hash, channel) pair with the interface.
// Received packets are returned to the channel whose hash matches the packet hash
func (l *LoopbackInterface) Register(hash uint64, c chan gopacket.Packet) {
	l.rxChannels[hash] = c
}

func (l *LoopbackInterface) Init() {}

func (l *LoopbackInterface) Send(p *gopacket.Packet) {
	l.txPkts++
	l.txBytes += uint64((*p).Metadata().CaptureInfo.CaptureLength)
	l.rxPkts = l.txPkts
	l.rxBytes = l.txBytes
	ch := l.rxChannels[(*p).NetworkLayer().NetworkFlow().FastHash()]
	ch <- *p
}

func (l *LoopbackInterface) TxStats() (txPkts, txBytes uint64) {
	return l.txPkts, l.txBytes
}

func (l *LoopbackInterface) RxStats() (rxPkts, rxBytes uint64) {
	return l.TxStats()  // By definition, Rx == Tx
}

func (l *LoopbackInterface) PktStats() (rxPkts, txPkts uint64) {
	return l.rxPkts, l.txPkts
}

func (l *LoopbackInterface) ByteStats() (rxBytes, txBytes uint64) {
	return l.rxBytes, l.txBytes
}

func (l *LoopbackInterface) Shutdown(timeout time.Duration) {}
