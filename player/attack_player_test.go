package player

import (
	"git.svc.rocks/dpariag/gotraffic/flow"
	"git.svc.rocks/dpariag/gotraffic/network"
	"git.svc.rocks/dpariag/gotraffic/stats"
	"net"
	"testing"
	"time"
)

func verifyAttackRate(attackFlowRate uint64, playSeconds uint64,
    stats stats.PlayerStats, t *testing.T) {

	expectedFlows := attackFlowRate * playSeconds
	// Allow 2% margin
	minExpectedFlows := uint64(float64(expectedFlows) * 0.98)
	maxExpectedFlows := uint64(float64(expectedFlows) * 1.02)

	if stats.FlowsCompleted < minExpectedFlows {
		t.Errorf("Sent %v attack flows. Expected at least %v flows\n",
			stats.FlowsCompleted, minExpectedFlows)
	}

	if stats.FlowsCompleted > maxExpectedFlows {
		t.Errorf("Sent %v attack flows. Expected at most %v flows\n",
			stats.FlowsCompleted, maxExpectedFlows)
	}
}

func TestSingleReflectorAttack(t *testing.T) {
	bridge := network.NewLoopbackBridgeGroup()
	flow := flow.NewFlow("../captures/quic.cap")
	ip := net.ParseIP("10.0.0.0")
	ports := []uint16{1}

	player := NewAttackPlayer(bridge, flow, ip, ports, 1)
	start := time.Now()
	player.PlayOnce()
	bridge.Shutdown(5 * time.Second)
	elapsed := time.Since(start)

	verifyPlayerStats(player, t)
	verifyFlowStats(player, flow, t)
	verifyBridgeStats(player, bridge, t)
	verifyReplayTime(elapsed, flow, 1, t)
}

func TestLowRateReflectorAttack(t *testing.T) {
	bridge := network.NewLoopbackBridgeGroup()
	flow := flow.NewFlow("../captures/quic.cap")
	ip := net.ParseIP("10.1.1.1")
	ports := []uint16{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	playSeconds := uint64(5)
	attackFlowRate := uint64(20)

	player := NewAttackPlayer(bridge, flow, ip, ports, attackFlowRate)
	player.Play()
	time.Sleep(time.Duration(playSeconds * 1000000000))
	player.Stop()
	bridge.Shutdown(5 * time.Second)

	verifyAttackRate(attackFlowRate, playSeconds, player.Stats(), t)
}

func TestHighRateReflectorAttack(t *testing.T) {
	bridge := network.NewLoopbackBridgeGroup()
	flow := flow.NewFlow("../captures/quic.cap")
	ip := net.ParseIP("10.1.1.1")
	ports := []uint16{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	playSeconds := uint64(5)
	attackFlowRate := uint64(1000)

	player := NewAttackPlayer(bridge, flow, ip, ports, attackFlowRate)
	player.Play()
	time.Sleep(time.Duration(playSeconds * 1000000000))
	player.Stop()
	bridge.Shutdown(5 * time.Second)

	verifyAttackRate(attackFlowRate, playSeconds, player.Stats(), t)
}
