package main

import (
	"time"
	"fmt"
	"git.svc.rocks/dpariag/gotraffic/flow"
	"git.svc.rocks/dpariag/gotraffic/player"
	"git.svc.rocks/dpariag/gotraffic/network"
)

func main() {
	bridge := network.NewLoopbackBridgeGroup()
	mix := flow.NewMix()
	//mix.AddFlow(flow.NewFlow("../captures/youtube-ip.cap"), 200)
	mix.AddFlow(flow.NewFlow("../captures/ping.cap"), 1)
	duration := 3 * time.Second

	p := player.NewMixPlayer(mix, bridge)
	p.Play()
	time.Sleep(duration)
	p.Stop()
	bridge.Shutdown(1*time.Second)

	bridgeStats := bridge.Stats()
	txPackets := bridgeStats.Client.Tx.Packets + bridgeStats.Server.Tx.Packets
	txBytes := bridgeStats.Client.Tx.Bytes + bridgeStats.Server.Tx.Bytes
	fmt.Printf("Device: TxPkts: %v, TxBytes: %v\n", txPackets, txBytes)
	fmt.Printf("Device: Pktrate: %v\n", float64(txPackets) / duration.Seconds())
	fmt.Printf("Device: Bitrate: %v\n", float64(txBytes * 8) / duration.Seconds())
}
