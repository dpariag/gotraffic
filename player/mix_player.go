package player

import (
	"git.svc.rocks/dpariag/gotraffic/flow"
	"git.svc.rocks/dpariag/gotraffic/network"
	"git.svc.rocks/dpariag/gotraffic/stats"
	"net"
)

//TODO: Should these be pointers (initialize by value too expensive?)
type MixPlayer struct {
	mix     flow.Mix            // The mix being played
	players []Player            // Players for each flow in the mix
	bridge  network.BridgeGroup // The bridge to write packets to
	ipGen   network.IPGenerator // Generate IPs for replay
}

func NewMixPlayer(m *flow.Mix, bridge network.BridgeGroup) *MixPlayer {
	mp := &MixPlayer{mix: *m, bridge: bridge, players: make([]Player, m.NumFlows()),
		ipGen: network.NewSequentialIPGenerator(net.ParseIP("10.0.0.1"))}

	flowNumber := 0
	for {
		fg, err := mp.mix.NextFlowGroup()
		if err != nil {
			break
		}
		for i := 0; i < int(fg.Copies); i++ {
			ips := mp.ipGen.GenerateIPs(2)
			mp.players[flowNumber] = NewPlayer(mp.bridge, &fg.Flow, ips)
			flowNumber++
		}
	}
	return mp
}

func (mp *MixPlayer) Stats() stats.PlayerStats {
	var curStats stats.PlayerStats
	for _, p := range mp.players {
		playerStats := p.Stats()
		curStats.Add(&playerStats)
	}
	return curStats
}

func (mp *MixPlayer) Play() {
	for _, p := range mp.players {
		p.Play()
	}
}

func (mp *MixPlayer) Stop() {
	for _, p := range mp.players {
		p.Stop()
	}
}
