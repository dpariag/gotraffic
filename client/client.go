package main

import (
	"fmt"
	"time"
	"git.svc.rocks/dpariag/gotraffic/flow"
	"git.svc.rocks/dpariag/gotraffic/network"
)

func main() {
	bridge := network.NewLoopbackBridgeGroup()
	mix := flow.NewMix()
	mix.AddFlow(flow.NewFlow("../captures/ping.cap"), 500)
	mix.AddFlow(flow.NewFlow("../captures/youtube-short.cap"), 50000)
	duration := 20 * time.Second

	player := flow.NewMixPlayer(mix, bridge)
	player.Play(duration)
	bridge.Shutdown(5*time.Second)

	bridgeStats := bridge.Stats()
	txPackets := bridgeStats.Client.Tx.Packets + bridgeStats.Server.Tx.Packets
	txBytes := bridgeStats.Client.Tx.Bytes + bridgeStats.Server.Tx.Bytes
	fmt.Printf("Device: TxPkts: %v, TxBytes: %v\n", txPackets, txBytes)
	fmt.Printf("Device: Bitrate: %v\n", float64(txBytes * 8) / duration.Seconds())
}

