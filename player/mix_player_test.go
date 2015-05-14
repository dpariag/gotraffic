package player

import (
	"git.svc.rocks/dpariag/gotraffic/flow"
	"git.svc.rocks/dpariag/gotraffic/network"
	"testing"
	"time"
)

func calculateFlowsToPlay(duration time.Duration, mix *flow.Mix) uint64 {
	var flowsToPlay uint64
	for fg, err := mix.NextFlowGroup(); err == nil; fg, err = mix.NextFlowGroup() {
		flowsPerDuration := uint64(((duration.Nanoseconds() / fg.Duration().Nanoseconds()) + 1))
		flowsToPlay += fg.Copies * flowsPerDuration
	}
	return flowsToPlay
}

func TestReplayBasicMix(t *testing.T) {
	bridge := network.NewLoopbackBridgeGroup()
	mix := flow.NewMix()
	mix.AddFlow(flow.NewFlow("../captures/ping.cap"), 1)
	mix.AddFlow(flow.NewFlow("../captures/youtube-short.cap"), 1)
	duration := 10 * time.Second
	flowsToPlay := calculateFlowsToPlay(duration, mix)

	player := NewMixPlayer(mix, bridge)
	player.Play()
	time.Sleep(duration)
	player.Stop()
	bridge.Shutdown(5 * time.Second)

	bridgeStats := bridge.Stats()
	playerStats := player.Stats()

	if playerStats.FlowsStarted != flowsToPlay {
		t.Errorf("Played %v flows in the mix. Should have played %v flows\n",
		playerStats.FlowsStarted, flowsToPlay)
	}

	if playerStats.FlowsCompleted > playerStats.FlowsStarted {
		t.Errorf("MixPlayer started %v flows, and completed %v flows\n",
			playerStats.FlowsStarted, playerStats.FlowsCompleted)
	}

	rxBytes := bridgeStats.Client.Rx.Bytes + bridgeStats.Server.Rx.Bytes
	if playerStats.Rx.Bytes != rxBytes {
		t.Errorf("MixPlayer rx: %v, Bridge rx: %v", playerStats.Rx.Bytes, rxBytes)
	}

	txBytes := bridgeStats.Client.Tx.Bytes + bridgeStats.Server.Tx.Bytes
	if playerStats.Tx.Bytes != txBytes {
		t.Errorf("MixPlayer tx: %v, Bridge tx: %v", playerStats.Tx.Bytes, txBytes)
	}
}
