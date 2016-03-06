// Implement the Player interface to send attack traffic
package player

import (
	"github.com/dpariag/gotraffic/flow"
	"github.com/dpariag/gotraffic/network"
	"net"
	"time"
)

type attackPlayer struct {
	playerCommon          // Embed the player basics
	flowBurst    uint64   // # of flows to burst
    sleepTime    uint64   // In nanoseconds
	dstIp        net.IP   // Destination IP
	dstPorts     []uint16 // Range of destination ports
	portIndex    uint64   // Index of current port
}

func NewAttackPlayer(bridge network.BridgeGroup, f *flow.Flow,
    ip net.IP, ports []uint16, flowRate uint64) Player {

	ap := attackPlayer{playerCommon: newPlayerCommon(bridge, f),
    dstIp: ip, dstPorts: ports, flowBurst: flowRate / uint64(10) }
    if ap.flowBurst == 0 {
        ap.flowBurst = 1
    }
    // Sleeping after every flow doesn't scale for single-packet flows
    // Instead, send a burst of flows (for now 1/10th of rate), then sleep.
    // Sleep time = (0.1s - time to send burst)
    ap.sleepTime = 100000000 - (ap.flowBurst * 20000 * f.NumPackets()) // Takes 20us to send a pkt
    ap.register()
	go ap.readPackets()
	return &ap
}

func (ap *attackPlayer) register() {
	dstRewrite := rewriteConfig{dstIp: ap.dstIp}
	firstPkt, err := newPacket(ap.flow.Packets()[0], dstRewrite)
	if err != nil {
		panic(err)
	}
	ap.bridge.Register(firstPkt.NetworkLayer().NetworkFlow().Dst(), ap.in)
}

func (ap *attackPlayer) playFlow() {
	dstRewrite := rewriteConfig{dstIp: ap.dstIp, dstPort: ap.dstPorts[ap.portIndex]}
	ap.portIndex = (ap.portIndex + 1) % uint64(len(ap.dstPorts))
	for _, p := range ap.flow.Packets() {
		// Clone the packet, re-writing the destination IP and port
		newPkt, err := newPacket(p, dstRewrite)
		if err != nil {
			panic(err)
		}
		ap.bridge.SendServerPacket(newPkt)
		ap.stats.Tx.Packets++
		ap.stats.Tx.Bytes += uint64(p.Packet.Metadata().CaptureInfo.CaptureLength)
		if ap.state == stopped {
			return
		}
	}
	ap.stats.FlowsCompleted++
}

func (ap *attackPlayer) Play() {
	ap.state = playing
	go func() {
        flowsSent := uint64(0)
		for ap.state != stopped {
			ap.playFlow()
            flowsSent++
            if flowsSent == ap.flowBurst {
                time.Sleep(time.Duration(ap.sleepTime))
                flowsSent = 0
            }
		}
	}()
}

func (ap *attackPlayer) PlayOnce() {
	ap.state = playing
	ap.playFlow()
	ap.state = stopped
}

func (ap *attackPlayer) Stop() {
	ap.state = stopped
}
