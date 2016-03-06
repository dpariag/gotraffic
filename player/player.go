package player

import (
	"fmt"
	"github.com/dpariag/gotraffic/flow"
	"github.com/dpariag/gotraffic/network"
	"github.com/google/gopacket"
	"net"
	"time"
)

type player struct {
	playerCommon          // Embed player basics
	ips          []net.IP // IPs that flows will be played from
	index        uint64   // Index of the ip currently in use
	state        int      // State of the player
}

func NewPlayer(bridge network.BridgeGroup, f *flow.Flow, ips []net.IP) Player {
	p := player{playerCommon: newPlayerCommon(bridge, f), ips: ips, state: stopped}

	for _, ip := range ips {
		p.register(ip)
	}
	go p.readPackets()
	return &p
}

// Registration is a bit of a hack
func (fp *player) register(subIp net.IP) {
	srcRewrite := rewriteConfig{srcIp: subIp}
	firstPkt, err := newPacket(fp.flow.Packets()[0], srcRewrite)
	if err != nil {
		panic(err)
	}
	fp.bridge.Register(firstPkt.NetworkLayer().NetworkFlow().Src(), fp.in)
}

func (fp *player) play() {
	// Choose an IP to play packets from
	srcRewrite := rewriteConfig{srcIp: fp.ips[fp.index]}
	dstRewrite := rewriteConfig{dstIp: fp.ips[fp.index]}
	//subIP := fp.ips[fp.index]
	fp.index = (fp.index + 1) % uint64(len(fp.ips))

	fp.stats.FlowsStarted++
	client := fp.flow.Client()
	// Write packets to the interface, respecting inter-packet gaps
	for _, p := range fp.flow.Packets() {
		// Clone the packet, re-writing the subscriber IP and
		// send it to correct side of the bridge group
		if p.Packet.NetworkLayer().NetworkFlow().Src() == client {
			newPkt, err := newPacket(p, srcRewrite)
			if err != nil {
				panic(err)
			}
			fp.bridge.SendClientPacket(newPkt)
		} else {
			newPkt, err := newPacket(p, dstRewrite)
			if err != nil {
				panic(err)
			}
			fp.bridge.SendServerPacket(newPkt)
		}
		fp.stats.Tx.Packets++
		fp.stats.Tx.Bytes += uint64(p.Packet.Metadata().CaptureInfo.CaptureLength)
		time.Sleep(p.Gap)
		if fp.state == stopped {
			return
		}
	}
	fp.stats.FlowsCompleted++
}

func (fp *player) Play() {
	fp.state = playing
	go func() {
		for fp.state != stopped {
			fp.play()
		}
	}()
}

func (fp *player) PlayOnce() {
	fp.state = playing
	fp.play()
	fp.state = stopped
}

func (fp *player) Stop() {
	fp.state = stopped
}

func printPacket(prefix string, p gopacket.Packet) {
	src, dst := p.NetworkLayer().NetworkFlow().Endpoints()
	fmt.Println(prefix, src, " --> ", dst)
}

func printFlow(flow *flow.Flow) {
	for _, p := range flow.Packets() {
		printPacket("", p)
	}
}
