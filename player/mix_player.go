package player

import (
	"git.svc.rocks/dpariag/gotraffic/flow"
	"git.svc.rocks/dpariag/gotraffic/network"
	"git.svc.rocks/dpariag/gotraffic/stats"
	"net"
)

// Associate a player with the index of it's flow in the Mix
type playerInfo struct {
	Player
	index int
}

//TODO: Should these be pointers (initialize by value too expensive?)
type MixPlayer struct {
	mix       flow.Mix            // The mix being played
	players   []playerInfo        // (Player,id) for each flow in the mix
	flowStats []stats.FlowStats   // Per-flow statistics
	bridge    network.BridgeGroup // The bridge to write packets to
	ipGen     network.IPGenerator // Generate IPs for replay
}

func NewMixPlayer(m *flow.Mix, bridge network.BridgeGroup) *MixPlayer {
	mp := &MixPlayer{mix: *m, bridge: bridge,
		players:   make([]playerInfo, m.NumFlows()),
		flowStats: make([]stats.FlowStats, m.NumFlowGroups()),
		ipGen:     network.NewSequentialIPGenerator(net.ParseIP("10.0.0.1"))}

	flowGroup, flowNumber := 0, 0
	flowGroups := mp.mix.FlowGroups()
	for index, fg := range flowGroups {
		for i := 0; i < int(fg.Copies); i++ {
			ips := mp.ipGen.GenerateIPs(2)
			mp.players[flowNumber] = playerInfo{NewPlayer(mp.bridge, flowGroups[index].Flow, ips), flowGroup}
			flowNumber++
		}
		mp.flowStats[flowGroup].Name = fg.Name()
		flowGroup++
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

func (mp *MixPlayer) FlowStats() []stats.FlowStats {
	for index, _ := range mp.flowStats {
		mp.flowStats[index].Clear()
	}

	for _, p := range mp.players {
		stat := p.Stats()
		mp.flowStats[p.index].Add(&stat)
	}
	return mp.flowStats
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
