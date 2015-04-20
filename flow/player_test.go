package flow

import (
	"git.svc.rocks/dpariag/gotraffic/network"
	"testing"
	"time"
)

// Estimate minimum and maximum acceptable replay times for a flow
// The calculated estimates are based on the number of packets in the flow
func flowReplayTolerances(f *Flow) (minReplayTime, maxReplayTime int64) {
	tolerance := int64(f.NumPkts() * 2000000) // 2ms per packet
	minReplayTime = f.Duration().Nanoseconds() - tolerance
	maxReplayTime = f.Duration().Nanoseconds() + tolerance
	return minReplayTime, maxReplayTime
}

// Check that the player received every packet it sent (no packet drops)
func verifyPlayerStats(p *Player, t *testing.T) {
	playerStats := p.Stats()

	if playerStats.Rx.Packets != playerStats.Tx.Packets {
		t.Errorf("Error: Player sent: %v packets, but received: %v packets\n",
				playerStats.Tx.Packets, playerStats.Rx.Packets)
	}

	if playerStats.Rx.Bytes != playerStats.Tx.Bytes {
		t.Errorf("Error: Player sent %v bytes, but received: %v bytes\n",
				playerStats.Tx.Bytes, playerStats.Rx.Bytes)
	}
}

// Check that the player sent packet count matches the flow packet count
func verifyFlowStats(p *Player, f *Flow, t *testing.T) {
	playerStats := p.Stats()

	if playerStats.Tx.Packets != f.NumPkts() {
		t.Errorf("Player sent %v pkts. Flow contains %v pkts\n", playerStats.Tx.Packets, f.NumPkts())
	}

	if playerStats.Tx.Bytes != f.NumBytes() {
		t.Errorf("Player sent %v bytes. Flow contains %v bytes\n", playerStats.Tx.Bytes, f.NumBytes())
	}
}

// Check that the player's sent and received packet counts match the interface
func verifyBridgeStats(p *Player, bg network.BridgeGroup, t *testing.T) {
	playerStats := p.Stats()
	bgStats := bg.Stats()

	bridgeTxPkts := bgStats.Client.Tx.Packets + bgStats.Server.Tx.Packets
	if playerStats.Tx.Packets != bridgeTxPkts {
		t.Errorf("Player TxPkts: %v Bridge TxPkts: %v\n", playerStats.Tx.Packets, bridgeTxPkts)
	}

	bridgeRxPkts := bgStats.Client.Rx.Packets + bgStats.Server.Rx.Packets
	if playerStats.Rx.Packets != bridgeRxPkts {
		t.Errorf("Player RxPkts: %v Bridge RxPkts: %v\n", playerStats.Rx.Packets, bridgeRxPkts)
	}
}

// Check that the time taken to replay the flow is acceptable
// Tolerate up to 2ms delay per packet
func verifyReplayTime(replayTime time.Duration, f *Flow, numReplays int64, t *testing.T) {
	minReplayTime, maxReplayTime := flowReplayTolerances(f)
	minReplayTime = minReplayTime * numReplays
	maxReplayTime = maxReplayTime * numReplays

	if replayTime.Nanoseconds() > maxReplayTime || replayTime.Nanoseconds() < minReplayTime {
		t.Errorf("replay time (ns): %v. number of replays: %v flow duration (ns): %v\n",
			replayTime, numReplays, f.Duration())
	}
}

func TestSingleFlowPlay(t *testing.T) {
	flow := NewFlow("../captures/ping.cap")
	bridge := network.NewLoopbackBridgeGroup()
	player := NewPlayer(bridge, flow)

	start := time.Now()
	player.Play()
	bridge.Shutdown(5 * time.Second)
	elapsed := time.Since(start)

	verifyPlayerStats(player, t)
	verifyFlowStats(player, flow, t)
	verifyBridgeStats(player, bridge, t)
	verifyReplayTime(elapsed, flow, 1, t)
}

func TestMultipleFlowReplay(t *testing.T) {
	flow := NewFlow("../captures/ping.cap")
	bridge := network.NewLoopbackBridgeGroup()
	player := NewPlayer(bridge, flow)
	done := make(chan *Player)

	start := time.Now()
	go player.Replay(done)
	player = <-done
	go player.Replay(done)
	<-done
	bridge.Shutdown(5 * time.Second)
	elapsed := time.Since(start)

	verifyPlayerStats(player, t)
	verifyBridgeStats(player, bridge, t)
	verifyReplayTime(elapsed, flow, 2, t)
}
