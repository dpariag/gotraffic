package player

import (
	"fmt"
	"git.svc.rocks/dpariag/gotraffic/flow"
	"git.svc.rocks/dpariag/gotraffic/network"
	"git.svc.rocks/dpariag/gotraffic/stats"
	"net"
	"sync"
	"time"
)

type MixPlayerStats struct {
	stats.Directional        // For bytes and packets
	flowsStarted      uint64 // Number of flows started during replay
	flowsCompleted    uint64 // Number of flows that have been completely replayed
}

//TODO: Should these be pointers (initialize by value too expensive?)
type MixPlayer struct {
	mix        flow.Mix            // The mix being played
	players    []*Player           // Players for each flow in the mix
	bridge     network.BridgeGroup // The bridge to write packets to
	ipGen      network.IPGenerator // Generate IPs for replay
	stats      MixPlayerStats      // Stats for the player
	statsLock  sync.Mutex          // Serialize concurrent updates
	replayChan chan *Player        // Completed players (can be restarted if necessary)
}

func NewMixPlayer(m *flow.Mix, bridge network.BridgeGroup) *MixPlayer {
	mp := &MixPlayer{mix: *m, bridge: bridge, players: make([]*Player, m.NumFlows()),
		ipGen:      network.NewSequentialIPGenerator(net.ParseIP("10.0.0.1")),
		replayChan: make(chan *Player, 10)}

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

func (mp *MixPlayer) Stats() MixPlayerStats {
	return mp.stats
}

func (mp *MixPlayer) Play(duration time.Duration) {
	start := time.Now()

	for _, p := range mp.players {
		mp.stats.flowsStarted++
		go p.Replay(mp.replayChan)
	}
	fmt.Printf("Started all players\n")
	for {
		replay := <-mp.replayChan
		mp.stats.flowsCompleted++

		if time.Since(start) > duration {
			mp.updateStats(replay)
			fmt.Printf("Time's up. Waiting for %v player to complete\n",
				mp.stats.flowsStarted-mp.stats.flowsCompleted)
			for mp.stats.flowsCompleted < mp.stats.flowsStarted {
				replay := <-mp.replayChan
				mp.stats.flowsCompleted++
				mp.updateStats(replay)
				fmt.Printf("Started: %v, Completed: %v\n",
					mp.stats.flowsStarted, mp.stats.flowsCompleted)
			}
			return
		}
		mp.playFlow(replay)
	}
}

func (mp *MixPlayer) playFlow(fp *Player) {
	mp.stats.flowsStarted++
	go fp.Replay(mp.replayChan)
}

func (mp *MixPlayer) updateStats(fp *Player) {
	mp.statsLock.Lock()
	playerStats := fp.Stats()
	mp.stats.Rx.Packets += playerStats.Rx.Packets
	mp.stats.Tx.Packets += playerStats.Tx.Packets
	mp.stats.Rx.Bytes += playerStats.Rx.Bytes
	mp.stats.Tx.Bytes += playerStats.Tx.Bytes
	mp.statsLock.Unlock()
}
