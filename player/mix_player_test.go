package flow

import (
	"git.svc.rocks/dpariag/gotraffic/flow"
	"git.svc.rocks/dpariag/gotraffic/network"
	"time"
	"testing"
)

func calculateFlowsToPlay(duration time.Duration, mix *flow.Mix) uint64 {
	var flowsToPlay uint64
	for fg,err := mix.NextFlowGroup(); err == nil; fg,err = mix.NextFlowGroup() {
		flowsPerDuration := uint64(((duration.Nanoseconds() / fg.Duration().Nanoseconds()) +  1))
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
	player.Play(duration)
	bridge.Shutdown(5*time.Second)

	bridgeStats := bridge.Stats()
	playerStats := player.Stats()

	if playerStats.flowsStarted != flowsToPlay {
		t.Errorf("Player should have played %v flows\n", flowsToPlay)
	}

	if playerStats.flowsStarted != playerStats.flowsCompleted {
		t.Errorf("Player only completed %v/%v flows\n",
				playerStats.flowsStarted, playerStats.flowsCompleted)
	}

	rxBytes := bridgeStats.Client.Rx.Bytes + bridgeStats.Server.Rx.Bytes
	if playerStats.Rx.Bytes != rxBytes {
		t.Errorf("Player rx: %v, Bridge rx: %v", playerStats.Rx.Bytes, rxBytes)
	}

	txBytes := bridgeStats.Client.Tx.Bytes + bridgeStats.Server.Tx.Bytes
	if playerStats.Tx.Bytes != txBytes {
		t.Errorf("Player tx: %v, Bridge tx: %v", playerStats.Tx.Bytes, txBytes)
	}
}
