package network

import (
	"fmt"
	"time"
	"code.google.com/p/gopacket"
	"code.google.com/p/gopacket/pcap"
)

func printPacket(prefix string, p *gopacket.Packet) {
	src,dst := (*p).NetworkLayer().NetworkFlow().Endpoints()
	fmt.Println(prefix, src, " --> ", dst)
}

type Device interface {
	Init()
	Register(hash uint64, c chan gopacket.Packet)
	Send(p *gopacket.Packet)
	PktStats() (rxPkts, txPkts uint64)
	ByteStats() (rxBytes, txBytes uint64)
	Shutdown(timeout time.Duration)
}

type PCAPInterface struct {
	handle *pcap.Handle
	txChan chan gopacket.Packet
	rxChannels map[uint64] chan gopacket.Packet

	rxPkts	uint64
	rxBytes	uint64
	txPkts	uint64
	txBytes	uint64
}

func NewPCAPInterface(name string, ) Device {
	i := PCAPInterface{}
	//TODO: Catch error
	i.handle, _ = pcap.OpenLive(name, 2048, true, pcap.BlockForever)
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
		//printPacket("W:", &p)
		i.handle.WritePacketData(p.Data())
		i.txPkts++
		i.txBytes += uint64(p.Metadata().CaptureInfo.CaptureLength)
	}
}

func (i *PCAPInterface) readPackets() {
	fmt.Println("Reader: Listening for packets.")
	packetSource := gopacket.NewPacketSource(i.handle, i.handle.LinkType())
	for p:= range packetSource.Packets() {
		//printPacket("R:", &p)
		i.rxPkts++
		i.rxBytes += uint64(p.Metadata().CaptureInfo.CaptureLength)
		ch := i.rxChannels[p.NetworkLayer().NetworkFlow().FastHash()]
		ch <- p
	}
}

func (i *PCAPInterface) Shutdown(timeout time.Duration) {
	start := time.Now()
	for i.rxPkts < i.txPkts && time.Since(start) < timeout {
		time.Sleep(time.Second)
	}
	fmt.Printf("Waited %v for packets to be returned\n", time.Since(start))
}

func (i *PCAPInterface) PktStats() (rxPkts, txPkts uint64) {
	return i.rxPkts, i.txPkts
}

func (i *PCAPInterface) ByteStats() (rxBytes, txBytes uint64) {
	return i.rxBytes, i.txBytes
}
