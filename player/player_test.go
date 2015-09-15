package player

import (
	"git.svc.rocks/dpariag/gotraffic/flow"
	"git.svc.rocks/dpariag/gotraffic/network"
	"net"
	"testing"
	"time"
)

func TestSingleFlowPlay(t *testing.T) {
	ip := []net.IP{net.ParseIP("10.0.0.0")}
	flow := flow.NewFlow("../captures/ping.cap")
	bridge := network.NewLoopbackBridgeGroup()
	player := NewPlayer(bridge, flow, ip)
	start := time.Now()
	player.PlayOnce()
	bridge.Shutdown(5 * time.Second)
	elapsed := time.Since(start)

	verifyPlayerStats(player, t)
	verifyFlowStats(player, flow, t)
	verifyBridgeStats(player, bridge, t)
	verifyReplayTime(elapsed, flow, 1, t)
}

func TestMultipleFlowReplay(t *testing.T) {
	ip := []net.IP{net.ParseIP("10.0.0.1"), net.ParseIP("10.0.0.1")}
	flow := flow.NewFlow("../captures/ping.cap")
	bridge := network.NewLoopbackBridgeGroup()
	player := NewPlayer(bridge, flow, ip)

	// Play the flow twice
	start := time.Now()
	player.PlayOnce()
	player.PlayOnce()
	bridge.Shutdown(5 * time.Second)
	elapsed := time.Since(start)

	verifyPlayerStats(player, t)
	verifyBridgeStats(player, bridge, t)
	verifyReplayTime(elapsed, flow, 2, t)
}

func TestPartialFlowReplay(t *testing.T) {
	flow := flow.NewFlow("../captures/ping.cap")
	playTime := 5 * time.Second
	if flow.Duration() < playTime {
		t.Errorf("Chosen flow is too short for partial replay test")
	}

	ip := []net.IP{net.ParseIP("10.0.0.2")}
	bridge := network.NewLoopbackBridgeGroup()
	player := NewPlayer(bridge, flow, ip)

	start := time.Now()
	player.Play()
	time.Sleep(playTime)
	player.Stop()
	elapsed := time.Since(start)
	bridge.Shutdown(5 * time.Second)

	verifyPlayerStats(player, t)
	verifyBridgeStats(player, bridge, t)
	playerStats := player.Stats()

	if playerStats.FlowsStarted != 1 {
		t.Errorf("Player started %v flows (should be one)", playerStats.FlowsStarted)
	}

	if playerStats.FlowsCompleted != 0 {
		t.Errorf("Player completed %v flows (should be zero)", playerStats.FlowsCompleted)
	}

	if elapsed > playTime+100*time.Millisecond {
		t.Errorf("Elapsed time is %v, should be %v\n", elapsed, playTime)
	}
}
