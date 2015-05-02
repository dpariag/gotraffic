package flow

import (
	"fmt"
	"time"
	"git.svc.rocks/dpariag/gotraffic/network"
	"git.svc.rocks/dpariag/gotraffic/stats"
	"github.com/google/gopacket"
)

type DevicePlayer struct {
	flow	Flow					// Flow being played
	device	network.Device			// Interface that packets are written to 
	in		chan gopacket.Packet	// channel that packets are returned on 
	stats	stats.Directional		// Rx and Tx stats
	done	chan *DevicePlayer			// Written on completion (allows easy replay)
}

func NewDevicePlayer(dev network.Device, f *Flow) *DevicePlayer {
	p := DevicePlayer{in:make(chan gopacket.Packet, len(f.pkts)),
				device:dev, flow:*f}
	p.device.Register(f.Hash(), p.in)
	return &p
}

func (fp *DevicePlayer) readPackets() {
	// TODO: Need timeout
	for i := 0; i < len(fp.flow.pkts); i++ {
		pkt := <-fp.in // read back the packet
		fp.stats.Rx.Packets++
		fp.stats.Rx.Bytes += uint64(pkt.Metadata().CaptureInfo.CaptureLength)
	}
}

func (fp *DevicePlayer) Play() {
	// Read packets from input channel
	go fp.readPackets()

	// Write packets to the interface, respecting inter-packet gaps
	start := time.Now()
	fmt.Printf("Starting...flow has %v pkts\n", len(fp.flow.pkts))
	for _, p := range fp.flow.pkts {
		// Send the embedded packet to the correct half of the bridge group
		if p.Packet.NetworkLayer().NetworkFlow().Src() == fp.flow.Client() {
			fp.device.Send(&p.Packet)
		} else {
			fp.device.Send(&p.Packet)
		}
		fp.stats.Tx.Packets++
		fp.stats.Tx.Bytes += uint64(p.Packet.Metadata().CaptureInfo.CaptureLength)
		time.Sleep(p.Gap)
	}
	fmt.Printf("Done flow playback. Elapsed: %v\n", time.Now().Sub(start))
}

// Play a flow, and signal completion by writing fp to the given channel upon completion
// This allows the channel owner to easily restart the flow
func (fp *DevicePlayer) Replay(done chan *DevicePlayer) {
	fp.Play()
	// TODO: Wait for all packets to be read back before signaling completion
	done <- fp
}

func (fp *DevicePlayer) Stats() stats.Directional {
	return fp.stats
}
