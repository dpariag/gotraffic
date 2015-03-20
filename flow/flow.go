package flow

import (
	"net"
	"time"
	"code.google.com/p/gopacket"
	"code.google.com/p/gopacket/pcap"
	"code.google.com/p/gopacket/layers"
)

type FlowPacket struct {
	gopacket.Packet		// The packet
	Gap	time.Duration	// Inter-packet gap between this packet and its predecessor
}

type FlowPackets []FlowPacket

type Flow struct {
	pkts		FlowPackets	// the packets of the flow, including inter-packet gaps
	duration	time.Duration
	numBytes	uint64
	bitrate		float64		// In bps
}

func NewFlow(capFile string) *Flow {
	if handle,err := pcap.OpenOffline(capFile); err != nil {
		panic(err)
	} else {
		var f Flow
		src:= gopacket.NewPacketSource(handle, handle.LinkType())
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

func (f *Flow) NumPkts() uint64 {
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

func (f Flow) Source() gopacket.Endpoint {
	return f.pkts[0].NetworkLayer().NetworkFlow().Src()
}

func (f Flow) Dest() gopacket.Endpoint {
	return f.pkts[0].NetworkLayer().NetworkFlow().Dst()
}

func (f *Flow) SourcePackets() FlowPackets {
	return f.endpointPackets(f.Source())
}

func (f *Flow) DestPackets() FlowPackets {
	return f.endpointPackets(f.Dest())
}

func (p *FlowPackets) RewriteIPs(srcIP net.IP, dstIP net.IP) {
	pkts := *p
	for i := 0; i < len(pkts); i++ {
		ip4 := pkts[i].NetworkLayer().(*layers.IPv4)
		ip4.SrcIP = srcIP
		ip4.DstIP = dstIP
	}
	*p = pkts
}

// TODOs
// 1) NewFlow should return an error instead of panic'ing 
// 2) Handle empty flow files

