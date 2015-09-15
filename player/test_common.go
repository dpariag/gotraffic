package player

import (
	"git.svc.rocks/dpariag/gotraffic/flow"
	"git.svc.rocks/dpariag/gotraffic/network"
	"testing"
	"time"
)

// TODO: Look into if/how go allows sharing of test code

// Estimate minimum and maximum acceptable replay times for a flow
// The calculated estimates are based on the number of packets in the flow
func flowReplayTolerances(f *flow.Flow) (minReplayTime, maxReplayTime int64) {
	tolerance := int64(f.NumPackets() * 2000000) // 2ms per packet
	minReplayTime = f.Duration().Nanoseconds() - tolerance
	maxReplayTime = f.Duration().Nanoseconds() + tolerance
	return minReplayTime, maxReplayTime
}

// Check that the player received every packet it sent (no packet drops)
func verifyPlayerStats(p Player, t *testing.T) {
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
func verifyFlowStats(p Player, f *flow.Flow, t *testing.T) {
	playerStats := p.Stats()

	if playerStats.Tx.Packets != f.NumPackets() {
		t.Errorf("Player sent %v pkts. Flow contains %v pkts\n",
			playerStats.Tx.Packets, f.NumPackets())
	}

	if playerStats.Tx.Bytes != f.NumBytes() {
		t.Errorf("Player sent %v bytes. Flow contains %v bytes\n", playerStats.Tx.Bytes, f.NumBytes())
	}
}

// Check that the player's sent and received packet counts match the interface
func verifyBridgeStats(p Player, bg network.BridgeGroup, t *testing.T) {
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
func verifyReplayTime(replayTime time.Duration, f *flow.Flow, numReplays int64, t *testing.T) {
	minReplayTime, maxReplayTime := flowReplayTolerances(f)
	minReplayTime = minReplayTime * numReplays
	maxReplayTime = maxReplayTime * numReplays

	if replayTime.Nanoseconds() > maxReplayTime {
		t.Errorf("replay time %v exceeds max replay time %v (ns)\n", replayTime, maxReplayTime)
	}
	if replayTime.Nanoseconds() < minReplayTime {
		t.Errorf("replay time %v is less than min replay time %v (ns)\n", replayTime, minReplayTime)
	}
}
