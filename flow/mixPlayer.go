package flow

import (
	"fmt"
	"git.svc.rocks/dpariag/gotraffic/network"
	"time"
	"sync"
)

type MixPlayerStats struct {
	network.DirectionalStats			// For bytes and packets
	flowsStarted	uint64				// Number of flows started during replay
	flowsCompleted	uint64				// Number of flows that have been completely replayed 
}

//TODO: Should these be pointers (initialize by value too expensive?)
type MixPlayer struct {
	mix				Mix					// The mix being played
	bridge			network.BridgeGroup	// The bridge to write packets to
	stats			MixPlayerStats		// Stats for the player
	statsLock		sync.Mutex			// Serialize concurrent updates
	replayChan		chan *Player		// Completed players (can be restarted if necessary)
}

func NewMixPlayer(m *Mix, bridge network.BridgeGroup) *MixPlayer {
	return &MixPlayer{mix: *m, bridge:bridge, replayChan:make(chan *Player, 10)}
}

func (mp *MixPlayer) Stats() MixPlayerStats {
	return mp.stats
}

func (mp *MixPlayer) Play(duration time.Duration) {
	start := time.Now()
	for {
		fg, err := mp.mix.NextFlowGroup()
		if err != nil {
			break
		}
		mp.playGroup(fg)
	}
	for {
		replay := <-mp.replayChan
		mp.stats.flowsCompleted++

		if time.Since(start) > duration {
			mp.updateStats(replay)
			fmt.Printf("Time's up. Waiting for %v player to complete\n", 
						mp.stats.flowsStarted - mp.stats.flowsCompleted)
			for mp.stats.flowsCompleted < mp.stats.flowsStarted {
				replay := <-mp.replayChan
				mp.stats.flowsCompleted++
				mp.updateStats(replay)
				fmt.Printf("Started: %v, Completed: %v\n", 
							mp.stats.flowsStarted, mp.stats.flowsCompleted)
			}
			return
		}
		fmt.Printf("restarting flow @  %v...\n", time.Now().Sub(start))
		mp.playFlow(replay)
	}

}

func (mp *MixPlayer) playGroup(f *FlowGroup) {
	// TODO: Each flow should be replayed with a different source IP 
	for i := 0; i < int(f.Copies); i++ {
		// TODO: Bug! NewPlayers can't Register() while other player are playing
		fp := NewPlayer(mp.bridge, &f.Flow)
		mp.playFlow(fp)
	}
}

func (mp *MixPlayer) playFlow(fp *Player) {
	mp.stats.flowsStarted++
	go fp.Replay(mp.replayChan)
}

func (mp *MixPlayer) updateStats(fp *Player) {
	mp.statsLock.Lock()
	playerStats := fp.Stats()
	fmt.Printf("Player stats: %v\n", playerStats.Rx.Bytes)
	mp.stats.Rx.Packets += playerStats.Rx.Packets
	mp.stats.Tx.Packets += playerStats.Tx.Packets
	mp.stats.Rx.Bytes += playerStats.Rx.Bytes
	mp.stats.Tx.Bytes += playerStats.Tx.Bytes
	mp.statsLock.Unlock()
}

