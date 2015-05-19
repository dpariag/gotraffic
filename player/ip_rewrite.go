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
	rewriteDest = 2
)

// newPacket changes the source or destination ip of the network layer
// of the given packet based on the from/to addresses. The metadata
// (capture info) is also updated, except for the timestamp, which is
// preserved from the original packet.
func newPacket(p gopacket.Packet, rewriteEndpoint int, to net.IP) (gopacket.Packet, error) {
	all := p.Layers()
	stack := make([]gopacket.SerializableLayer, len(all))
	var networkLayer gopacket.NetworkLayer
	for n, layer := range all {
		// TODO(afiori): Handle IP fragments?
		switch layer.LayerType() {
		case layers.LayerTypeIPv4:
			ip4 := layer.(*layers.IPv4)
			if rewriteEndpoint == rewriteSource {
				ip4.SrcIP = to.To4()
			} else {
				ip4.DstIP = to.To4()
			}
			stack[n] = ip4
			networkLayer = ip4
		case layers.LayerTypeIPv6:
			// TODO: Test v6.
			ip6 := layer.(*layers.IPv6)
			if rewriteEndpoint == rewriteSource {
				ip6.SrcIP = to
			} else {
				ip6.DstIP = to
			}
			stack[n] = ip6
			networkLayer = ip6
		case layers.LayerTypeTCP:
			if networkLayer != nil {
				tcp := layer.(*layers.TCP)
				tcp.SetNetworkLayerForChecksum(networkLayer)
				stack[n] = tcp
			}
		case layers.LayerTypeUDP:
			if networkLayer != nil {
				udp := layer.(*layers.UDP)
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
