package network

import (
	"git.svc.rocks/dpariag/gotraffic/stats"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"time"
)

const (
	clientDevice = 1
	serverDevice = 2
)

type ioHandle interface {
	ReadPacketData() (data []byte, ci gopacket.CaptureInfo, err error)
	WritePacketData(data []byte) (err error)
}

type Device interface {
	Register(ep gopacket.Endpoint, c chan gopacket.Packet)
	Send(p gopacket.Packet)
	Stats() stats.Directional
	Shutdown(timeout time.Duration)
}

type BridgeGroup interface {
	Register(ep gopacket.Endpoint, c chan gopacket.Packet)
	SendClientPacket(p gopacket.Packet)
	SendServerPacket(p gopacket.Packet)
	Stats() stats.BridgeGroupStats
	Shutdown(timeout time.Duration)
}

type device struct {
	role       int                             // Role in a bridge group (client side or server side)
	io         ioHandle                        // For reading/writing packets
	stats      stats.Directional               // Device stats
	txChan     chan gopacket.Packet            // Packets to be sent are collected from here
	rxChannels map[uint64]chan gopacket.Packet // Each packet is returned on a channel
}

// Subtle: This should return a pointer!
func newDevice(name string, role int, io ioHandle) *device {
	d := device{role: role, io: io, txChan: make(chan gopacket.Packet, 10),
		rxChannels: make(map[uint64]chan gopacket.Packet)}
	d.init()
	return &d
}

// Register a (endpoint, channel) pair with the interface.
// Received packets are returned to the channel whose hash matches the endpoint hash
func (d *device) Register(ep gopacket.Endpoint, c chan gopacket.Packet) {
	d.rxChannels[ep.FastHash()] = c
}

func (i *device) init() {
	go i.readPackets()
	go i.sendPackets()
}

func (d *device) Send(p gopacket.Packet) {
	d.txChan <- p
}

func (d *device) sendPackets() {
	for {
		p := (<-d.txChan)
		d.io.WritePacketData(p.Data())
		d.stats.Tx.Packets++
		d.stats.Tx.Bytes += uint64(p.Metadata().CaptureInfo.CaptureLength)
	}
}

func (d *device) readPackets() {
	var hash uint64
	for true {
		data, ci, err := d.io.ReadPacketData()
		if err != nil {
			break
		}

		d.stats.Rx.Packets++
		d.stats.Rx.Bytes += uint64(ci.CaptureLength)
		p := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.Default)
		if d.role == clientDevice {
			hash = p.NetworkLayer().NetworkFlow().Dst().FastHash()
		} else {
			hash = p.NetworkLayer().NetworkFlow().Src().FastHash()
		}
		ch := d.rxChannels[hash]
		if ch != nil {
			ch <- p
		}
	}
}

func (d *device) Shutdown(timeout time.Duration) {
	start := time.Now()
	for d.stats.Rx.Packets != d.stats.Tx.Packets && time.Since(start) < timeout {
		time.Sleep(time.Millisecond * 50)
	}
}

func (d *device) Stats() stats.Directional {
	return d.stats
}
