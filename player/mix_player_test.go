package player

import (
	"github.com/dpariag/gotraffic/flow"
	"github.com/dpariag/gotraffic/network"
	"github.com/dpariag/gotraffic/stats"
	"testing"
	"time"
)

func calculateFlowsToPlay(duration time.Duration, mix *flow.Mix) uint64 {
	var flowsToPlay uint64
	for _,fg := range mix.FlowGroups() {
		flowsPerDuration := uint64(((duration.Nanoseconds() / fg.Duration().Nanoseconds()) + 1))
		flowsToPlay += fg.Copies * flowsPerDuration
	}
	return flowsToPlay
}

// Calculate how many flows should be started and completed in a given duration
func calculateFlowStats(duration time.Duration, f *flow.Flow, copies uint64) (started uint64, completed uint64) {
	flowDuration := f.Duration().Nanoseconds()
	completed = uint64(duration.Nanoseconds() / flowDuration)
	return (completed + 1) * copies, completed * copies
}

func findFlowStat(flowName string, flowStats []stats.FlowStats) *stats.FlowStats {
	for i, _ := range flowStats {
		if flowStats[i].Name == flowName {
			return &flowStats[i]
		}
	}
	return nil
}

func verifyPerFlowStats(flowStat *stats.FlowStats, started uint64, completed uint64, t *testing.T) {
	if flowStat.FlowsStarted != started {
		t.Errorf("Started %v %s flows. Should have started %v\n",
			flowStat.FlowsStarted, flowStat.Name, started)
	}

	if flowStat.FlowsCompleted != completed {
		t.Errorf("Completed %v %s flows. Should have completed %v\n",
			flowStat.FlowsCompleted, flowStat.Name, completed)
	}
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
		t.Errorf("Started %v flows in the mix. Should have started %v flows\n",
			playerStats.FlowsStarted, flowsToPlay)
	}

	if playerStats.FlowsCompleted > playerStats.FlowsStarted {
		t.Errorf("MixPlayer started %v flows, but completed %v flows\n",
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

func TestPerFlowStats(t *testing.T) {
	bridge := network.NewLoopbackBridgeGroup()
	http := flow.NewFlow("../captures/http.cap")
	ping := flow.NewFlow("../captures/ping.cap")

	mix := flow.NewMix()
	mix.AddFlow(ping, 2)
	mix.AddFlow(http, 5)
	duration := 16 * time.Second
	httpStarted, httpCompleted := calculateFlowStats(duration, http, 5)
	pingStarted, pingCompleted := calculateFlowStats(duration, ping, 2)

	player := NewMixPlayer(mix, bridge)
	player.Play()
	time.Sleep(duration)
	player.Stop()
	bridge.Shutdown(5 * time.Second)

	flowStats := player.FlowStats()
	httpStat := findFlowStat(http.Name(), flowStats)
	pingStat := findFlowStat(ping.Name(), flowStats)

	verifyPerFlowStats(httpStat, httpStarted, httpCompleted, t)
	verifyPerFlowStats(pingStat, pingStarted, pingCompleted, t)
}
