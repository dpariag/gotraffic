package flow

import (
	"fmt"
	"time"
	"github.com/dpariag/network"
)

type MixPlayer struct {
	mix				Mix					// The mix being played
	iface			network.Interface	// The interface to write packets to (TODO: Need 2)
	duration		time.Duration		// Duration of traffic replay
	flowsStarted	uint64				// Number of flows started during replay
	flowsCompleted	uint64				// Number of flows that have been completely replayed 
	replayChan		chan *Player		// Completed players (can be restarted if necessary)
}

func NewMixPlayer(m *Mix, iface network.Interface, duration time.Duration) *MixPlayer {
	return &MixPlayer{mix: *m, iface:iface, duration:duration, replayChan:make(chan *Player, 10)}
}

func (mp *MixPlayer) Play() {
	start := time.Now()
	for {
		fg, err := mp.mix.NextFlowGroup();
		if err != nil {
			break
		}
		mp.playGroup(fg)
	}
	for {
		replay := <-mp.replayChan
		mp.flowsCompleted++

		if time.Since(start) > mp.duration {
			print("****Time's up. Waiting for existing flows to complete\n")
			for mp.flowsCompleted < mp.flowsStarted {
				<-mp.replayChan
				mp.flowsCompleted++
				fmt.Printf("Started: %v, Completed: %v\n", mp.flowsStarted, mp.flowsCompleted)
			}
			return
		}
		fmt.Printf("restarting flow @  %v...\n", time.Now().Sub(start))
		go mp.playFlow(replay)
	}

}

func (mp *MixPlayer) playGroup(f *FlowGroup) {
	// TODO: Each flow should be replayed with a different 5-tuple
	for i := 0; i < int(f.Copies); i++ {
		fp := NewPlayer(mp.iface, &f.Flow)
		go mp.playFlow(fp)
	}
}

func (mp *MixPlayer) playFlow(fp *Player) {
	mp.flowsStarted++
	fp.Replay(mp.replayChan)
}
