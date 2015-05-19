package network

import (
	"git.svc.rocks/dpariag/gotraffic/stats"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"time"
)


type pcapDevice struct {
	role	   int
	handle     *pcap.Handle
	txChan     chan gopacket.Packet
	rxChannels map[uint64]chan gopacket.Packet
	stats      stats.Directional
}

func NewPCAPDevice(name string, role int) Device {
	i := pcapDevice{role:role, txChan:make(chan gopacket.Packet, 10),
	rxChannels:make(map[uint64]chan gopacket.Packet) }

	handle, err := pcap.OpenLive(name, 2048, true, pcap.BlockForever)
	if err != nil {
		panic(err)
	}
	i.handle = handle
	i.init()
	return &i
}

// Register a (endpoint, channel) pair with the interface.
// Received packets are returned to the channel whose hash matches the endpoint hash
func (i *pcapDevice) Register(ep gopacket.Endpoint, c chan gopacket.Packet) {
	i.rxChannels[ep.FastHash()] = c
}

func (i *pcapDevice) init() {
	go i.readPackets()
	go i.sendPackets()
}

func (i *pcapDevice) Send(p gopacket.Packet) {
	i.txChan <- p
}

func (i *pcapDevice) sendPackets() {
	for {
		p := (<-i.txChan)
		i.handle.WritePacketData(p.Data())
		i.stats.Tx.Packets++
		i.stats.Tx.Bytes += uint64(p.Metadata().CaptureInfo.CaptureLength)
	}
}

func (i *pcapDevice) readPackets() {
	packetSource := gopacket.NewPacketSource(i.handle, i.handle.LinkType())
	var hash uint64
	for p := range packetSource.Packets() {
		i.stats.Rx.Packets++
		i.stats.Rx.Bytes += uint64(p.Metadata().CaptureInfo.CaptureLength)
		if i.role == clientDevice { 
			hash = p.NetworkLayer().NetworkFlow().Dst().FastHash()
		} else {
			hash = p.NetworkLayer().NetworkFlow().Src().FastHash()
		}
		ch := i.rxChannels[hash]
		if ch != nil {
			ch <- p
		}
	}
}

func (i *pcapDevice) Shutdown(timeout time.Duration) {
	start := time.Now()
	for i.stats.Rx.Packets < i.stats.Tx.Packets && time.Since(start) < timeout {
		time.Sleep(time.Second)
	}
}

func (i *pcapDevice) Stats() stats.Directional {
	return i.stats
}
