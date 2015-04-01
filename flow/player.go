package flow

import (
	"code.google.com/p/gopacket"
	"fmt"
	"git.svc.rocks/dpariag/gotraffic/network"
	"time"
)

type Player struct {
	flow   Flow                 // Flow being played
	iface  network.Interface    // Interface that packets are written to
	in     chan gopacket.Packet // channel that packets are returned on
	rxPkts uint64               // Num of packets received from in
	txPkts uint64               // Num of packets sent to out
	done   chan *Player         // Written on completion (allows easy replay)
}

func NewPlayer(iface network.Interface, f *Flow) *Player {
	p := Player{in: make(chan gopacket.Packet, len(f.pkts)),
		iface: iface, flow: *f, rxPkts: 0, txPkts: 0}
	p.iface.Register(f.Hash(), p.in)
	return &p
}

func (fp *Player) readPackets() {
	// TODO: Need timeout
	for i := 0; i < len(fp.flow.pkts); i++ {
		<-fp.in // read back the packet
		fp.rxPkts++
	}
}

// Play a flow
func (fp *Player) Play() {
	// Read packets from input channel
	go fp.readPackets()

	// Send packets to output channel, respecting inter-packet gaps
	start := time.Now()
	fmt.Printf("Starting...flow has %v pkts\n", len(fp.flow.pkts))
	for _, p := range fp.flow.pkts {
		// Send, referencing the embedded packet
		fp.iface.Send(&p.Packet)
		fp.txPkts++
		time.Sleep(p.Gap)
	}
	fmt.Printf("Done flow playback. Elapsed: %v\n", time.Now().Sub(start))
}

// Play a flow, and signal completion by writing fp to the given channel upon completion
// This allows the channel owner to easily restart the flow
func (fp *Player) Replay(done chan *Player) {
	fp.Play()
	// TODO: Wait for all packets to be read back before signaling completion
	done <- fp
}

func (fp *Player) PktStats() (rxPkts, txPkts uint64) {
	return fp.rxPkts, fp.txPkts
}

func (fp *Player) DroppedPkts() uint64 {
	// Only report non-zero drops if we're done playing the flow
	if fp.txPkts < uint64(len(fp.flow.pkts)) {
		return 0
	}
	return (fp.txPkts - fp.rxPkts)
}
