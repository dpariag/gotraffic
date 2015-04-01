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
	playerRxPkts, playerTxPkts := p.PktStats()

	if playerRxPkts != playerTxPkts {
		t.Errorf("Error: Player sent: %v packets, but received: %v packets\n", playerTxPkts, playerRxPkts)
	}

	if p.DroppedPkts() != 0 {
		t.Errorf("Error: Flow player reports %v dropped packets\n", p.DroppedPkts())
	}
}

// Check that the player sent packet count matches the flow packet count
func verifyFlowStats(p *Player, f *Flow, t *testing.T) {
	_, playerTxPkts := p.PktStats()
	flowPkts := f.NumPkts()
	if playerTxPkts != flowPkts {
		t.Errorf("Player sent %v pkts. Flow contains %v pkts\n", playerTxPkts, flowPkts)
	}
}

// TODO: i should be network.Device
// Check that the player's sent and received packet counts match the interface
func verifyInterfaceStats(p *Player, i network.Device, t *testing.T) {
	playerRxPkts, playerTxPkts := p.PktStats()
	ifaceRxPkts, ifaceTxPkts := i.PktStats()

	if playerTxPkts != ifaceTxPkts {
		t.Errorf("Player TxPkts: %v Interface TxPkts: %v\n", playerTxPkts, ifaceTxPkts)
	}

	if playerRxPkts != ifaceRxPkts {
		t.Errorf("Player RxPkts: %v Interface RxPkts: %v\n", playerRxPkts, ifaceRxPkts)
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
	flow := NewFlow("captures/ping.cap")
	iface := network.NewLoopback()
	iface.Init()
	player := NewPlayer(iface, flow)

	start := time.Now()
	player.Play()
	iface.Shutdown(5 * time.Second)
	elapsed := time.Since(start)

	verifyPlayerStats(player, t)
	verifyFlowStats(player, flow, t)
	verifyInterfaceStats(player, iface, t)
	verifyReplayTime(elapsed, flow, 1, t)
}

func TestMultipleFlowReplay(t *testing.T) {
	flow := NewFlow("captures/ping.cap")
	iface := network.NewLoopback()
	iface.Init()
	player := NewPlayer(iface, flow)
	done := make(chan *Player)

	start := time.Now()
	go player.Replay(done)
	<-done
	go player.Replay(done)
	<-done
	iface.Shutdown(5 * time.Second)
	elapsed := time.Since(start)

	verifyPlayerStats(player, t)
	verifyInterfaceStats(player, iface, t)
	verifyReplayTime(elapsed, flow, 2, t)
}
