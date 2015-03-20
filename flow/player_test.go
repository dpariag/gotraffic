package flow

import (
	"fmt"
	"time"
	"testing"
	"gotraffic/network"
)

func flowReplayTolerances(f *Flow) (minReplayTime, maxReplayTime int64) {
	tolerance := int64(f.NumPkts() * 2000000)
	minReplayTime = f.Duration().Nanoseconds() - tolerance
	maxReplayTime = f.Duration().Nanoseconds() + tolerance
	fmt.Printf("Min: %v Max: %v\n", minReplayTime, maxReplayTime)
	return minReplayTime, maxReplayTime
}

func TestSingleFlowReplay(t *testing.T) {
	flow := NewFlow("captures/ping.cap")
	iface := network.NewLoopback()
	iface.Init()
	player := NewPlayer(iface, flow)

	start := time.Now()
	player.Play()
	iface.Shutdown(5*time.Second)
	elapsed := time.Since(start)

	flowPkts := flow.NumPkts()
	playerRxPkts, playerTxPkts := player.PktStats()
	ifaceRxPkts, ifaceTxPkts := iface.PktStats()

	if playerRxPkts != playerTxPkts {
		t.Errorf("Error: Player txPkts: %v rxPkts: %v\n", playerRxPkts, playerTxPkts)
	}

	if player.DroppedPkts() != 0 {
		t.Errorf("Error: Flow player reports %v dropped packets\n", player.DroppedPkts())
	}

	if playerTxPkts != flowPkts {
		t.Errorf("Player sent %v pkts. Flow contains %v pkts\n", playerTxPkts, flowPkts)
	}

	if playerTxPkts != ifaceTxPkts {
		t.Errorf("Player TxPkts: %v Interface TxPkts: %v\n", playerTxPkts, ifaceTxPkts)
	}

	if playerRxPkts != ifaceRxPkts {
		t.Errorf("Player RxPkts: %v Interface RxPkts: %v\n", playerRxPkts, ifaceRxPkts)
	}

	minReplayTime, maxReplayTime := flowReplayTolerances(flow)
	if elapsed.Nanoseconds() > maxReplayTime || elapsed.Nanoseconds() < minReplayTime {
		t.Errorf("Replay time: %v. Actual flow duration: %v\n", elapsed, flow.Duration())
	}
}
