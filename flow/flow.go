package flow

import (
	"time"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

type FlowPacket struct {
	gopacket.Packet               // The packet
	Gap             time.Duration // Inter-packet gap between this packet and its predecessor
}

type FlowPackets []FlowPacket

type Flow struct {
	pkts     FlowPackets // the packets of the flow, including inter-packet gaps
	duration time.Duration
	numBytes uint64
	bitrate  float64 // In bps
}

func NewFlow(capFile string) *Flow {
	if handle, err := pcap.OpenOffline(capFile); err != nil {
		panic(err)
	} else {
		var f Flow
		src := gopacket.NewPacketSource(handle, handle.LinkType())
		f.pkts = make(FlowPackets, 0)
		lastPktTime := time.Now()

		for pkt := range src.Packets() {
			pktTime := pkt.Metadata().CaptureInfo.Timestamp
			f.pkts = append(f.pkts, FlowPacket{pkt, pktTime.Sub(lastPktTime)})
			lastPktTime = pktTime
			f.numBytes += uint64(pkt.Metadata().CaptureInfo.CaptureLength)
		}

		f.pkts[0].Gap = 0 // Fix the gap on the first packet
		f.setDuration()
		f.bitrate = float64(f.numBytes) * 8.0 / f.duration.Seconds()
		return &f
	}
}

func (f *Flow) setDuration() {
	startTime := f.pkts[0].Metadata().CaptureInfo.Timestamp
	endTime := f.pkts[len(f.pkts)-1].Metadata().CaptureInfo.Timestamp
	f.duration = endTime.Sub(startTime)
}

func (f *Flow) Packets() FlowPackets {
	return f.pkts
}

func (f *Flow) NumPackets() uint64 {
	return uint64(len(f.pkts))
}

func (f *Flow) NumBytes() uint64 {
	return f.numBytes
}

func (f *Flow) Duration() time.Duration {
	return f.duration
}

func (f *Flow) Bitrate() float64 {
	return f.bitrate
}

func (f Flow) Hash() uint64 {
	return f.pkts[0].NetworkLayer().NetworkFlow().FastHash()
}

func (f *Flow) endpointPackets(endpoint gopacket.Endpoint) FlowPackets {
	pkts := make(FlowPackets, 0)
	for i := 0; i < len(f.pkts); i++ {
		if f.pkts[i].NetworkLayer().NetworkFlow().Src() == endpoint {
			pkts = append(pkts, f.pkts[i])
		}
	}
	return pkts
}

// Client (initiator) is the source of the first packet
func (f *Flow) Client() gopacket.Endpoint {
	return f.pkts[0].NetworkLayer().NetworkFlow().Src()
}

// Server (recipient) is the dest of the first packet
func (f *Flow) Server() gopacket.Endpoint {
	return f.pkts[0].NetworkLayer().NetworkFlow().Dst()
}

func (f *Flow) ClientPackets() FlowPackets {
	return f.endpointPackets(f.Client())
}

func (f *Flow) ServerPackets() FlowPackets {
	return f.endpointPackets(f.Server())
}

func printPacket(prefix string, p gopacket.Packet) {
	src, dst := p.NetworkLayer().NetworkFlow().Endpoints()
	fmt.Println(prefix, src, " --> ", dst)
}

/*
//This modifies the gopacket layers, but not the underlying packet data
func (f *Flow) rewriteIPs(srcIP net.IP, dstIP net.IP) {
	for i := 0; i < len(f.pkts); i++ {
		ip4 := f.pkts[i].NetworkLayer().(*layers.IPv4)
		ip4.SrcIP = srcIP
		ip4.DstIP = dstIP
		printPacket("Rewrite:", f.pkts[i].Packet)
	}
}

func (f *Flow) rewriteSourceIP(srcIp net.IP) FlowPackets {
	pkts := make(FlowPackets, len(f.pkts))
	for i := 0; i < len(f.pkts); i++ {
		pkt,err := newPacket(f.pkts[i].Packet, srcIp)
		if err != nil {
			panic(err)
		}
		pkts[i] = FlowPacket{pkt, pkts[i].Gap}
	}
	return pkts
}

func (f *Flow) Clone(srcIp net.IP) Flow {
	return Flow{f.rewriteSourceIP(srcIp), f.duration, f.numBytes, f.bitrate}
}
*/
