// Based on code from afiori

package player

import (
	"errors"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"net"
	"time"
)

const (
	rewriteSource = 1
	rewriteDest   = 2
)

type rewriteConfig struct {
	srcIp   net.IP // nil indicates no rewrite necessary
	dstIp   net.IP // nil indicates no rewrite necessary
	srcPort uint16 // 0 indicates no rewrite necessary
	dstPort uint16 // 0 indicates no rewrite necessary
}

// newPacket changes the source or destination ip of the network layer
// of the given packet based on the from/to addresses. The metadata
// (capture info) is also updated, except for the timestamp, which is
// preserved from the original packet.
func newPacket(p gopacket.Packet, rewrite rewriteConfig) (gopacket.Packet, error) {
	all := p.Layers()
	stack := make([]gopacket.SerializableLayer, len(all))
	var networkLayer gopacket.NetworkLayer
	for n, layer := range all {
		// TODO(afiori): Handle IP fragments?
		switch layer.LayerType() {
		case layers.LayerTypeIPv4:
			// Copy the IPv4 layer
			ip4 := *layer.(*layers.IPv4)
			if rewrite.srcIp != nil {
				ip4.SrcIP = rewrite.srcIp.To4()
			}
			if rewrite.dstIp != nil {
				ip4.DstIP = rewrite.dstIp.To4()
			}
			stack[n] = &ip4
			networkLayer = &ip4
		case layers.LayerTypeIPv6:
			// TODO: Test v6.
			ip6 := layer.(*layers.IPv6)
			if rewrite.srcIp != nil {
				ip6.SrcIP = rewrite.srcIp
			}
			if rewrite.dstIp != nil {
				ip6.DstIP = rewrite.dstIp
			}
			stack[n] = ip6
			networkLayer = ip6
		case layers.LayerTypeTCP:
			if networkLayer != nil {
				tcp := layer.(*layers.TCP)
				if rewrite.srcPort != 0 {
					tcp.SrcPort = layers.TCPPort(rewrite.srcPort)
				}
				if rewrite.dstPort != 0 {
					tcp.DstPort = layers.TCPPort(rewrite.dstPort)
				}
				tcp.SetNetworkLayerForChecksum(networkLayer)
				stack[n] = tcp
			}
		case layers.LayerTypeUDP:
			if networkLayer != nil {
				udp := layer.(*layers.UDP)
				if rewrite.srcPort != 0 {
					udp.SrcPort = layers.UDPPort(rewrite.srcPort)
				}
				if rewrite.dstPort != 0 {
					udp.DstPort = layers.UDPPort(rewrite.dstPort)
				}
				udp.SetNetworkLayerForChecksum(networkLayer)
				stack[n] = udp
			}
		case layers.LayerTypeDNS:
			// Stupid DNS decoding based on UDP port.
			dns := layer.(*layers.DNS)
			stack[n] = gopacket.Payload(dns.BaseLayer.Contents)
		default:
			if s, ok := layer.(gopacket.SerializableLayer); ok {
				stack[n] = s
			}
		}
	}
	m := p.Metadata()
	return buildPacket(m.CaptureInfo.Timestamp, stack)
}

// buildPacket creates a new gopacket.Packet from the given stack of
// network layers, and sets the metadata correctly so this packet can
// be written to a pcap file.
func buildPacket(ts time.Time, stack []gopacket.SerializableLayer) (gopacket.Packet, error) {
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}
	err := gopacket.SerializeLayers(buf, opts, stack...)
	if err != nil {
		return nil, err
	}
	dec := gopacket.DecodeOptions{NoCopy: true, Lazy: true}
	p := gopacket.NewPacket(buf.Bytes(), layers.LayerTypeEthernet, dec)
	if p == nil {
		return nil, errors.New("failed to build packet from new data")
	}
	m := p.Metadata()
	l := len(p.Data())
	m.CaptureInfo.CaptureLength = l
	m.CaptureInfo.Length = l
	m.CaptureInfo.Timestamp = ts
	return p, nil
}
