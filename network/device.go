package network

import (
	"fmt"
	"git.svc.rocks/dpariag/gotraffic/stats"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"time"
)

func printPacket(prefix string, p *gopacket.Packet) {
	src, dst := (*p).NetworkLayer().NetworkFlow().Endpoints()
	fmt.Println(prefix, src, " --> ", dst)
}

type Device interface {
	Init()
	Register(hash uint64, c chan gopacket.Packet)
	Send(p *gopacket.Packet)
	Stats() stats.Directional
	Shutdown(timeout time.Duration)
}

type PCAPInterface struct {
	handle     *pcap.Handle
	txChan     chan gopacket.Packet
	rxChannels map[uint64]chan gopacket.Packet
	stats      stats.Directional
}

func NewPCAPInterface(name string) Device {
	i := PCAPInterface{}
	handle, err := pcap.OpenLive(name, 2048, true, pcap.BlockForever)
	if err != nil {
		panic(err)
	}
	i.handle = handle
	i.txChan = make(chan gopacket.Packet, 10)
	i.rxChannels = make(map[uint64]chan gopacket.Packet)
	return &i
}

// Register a (hash, channel) pair with the interface.
// Received packets are returned to the channel whose hash matches the packet hash
func (i *PCAPInterface) Register(hash uint64, c chan gopacket.Packet) {
	i.rxChannels[hash] = c
}

func (i *PCAPInterface) Init() {
	go i.readPackets()
	go i.sendPackets()
}

func (i *PCAPInterface) Send(p *gopacket.Packet) {
	i.txChan <- *p
}

func (i *PCAPInterface) sendPackets() {
	fmt.Printf("Writer: Waiting for a packet...")
	for {
		p := (<-i.txChan)
		i.handle.WritePacketData(p.Data())
		i.stats.Tx.Packets++
		i.stats.Tx.Bytes += uint64(p.Metadata().CaptureInfo.CaptureLength)
	}
}

func (i *PCAPInterface) readPackets() {
	fmt.Println("Reader: Listening for packets.")
	packetSource := gopacket.NewPacketSource(i.handle, i.handle.LinkType())
	for p := range packetSource.Packets() {
		i.stats.Rx.Packets++
		i.stats.Rx.Bytes += uint64(p.Metadata().CaptureInfo.CaptureLength)
		ch := i.rxChannels[p.NetworkLayer().NetworkFlow().FastHash()]
		ch <- p
	}
}

func (i *PCAPInterface) Shutdown(timeout time.Duration) {
	start := time.Now()
	for i.stats.Rx.Packets < i.stats.Tx.Packets && time.Since(start) < timeout {
		time.Sleep(time.Second)
	}
	fmt.Printf("Waited %v for packets to be returned\n", time.Since(start))
}

func (i *PCAPInterface) Stats() stats.Directional {
	return i.stats
}
