package main

import (
	"fmt"
	"git.svc.rocks/dpariag/gotraffic/flow"
	"git.svc.rocks/dpariag/gotraffic/network"
	"time"
)

func main() {
	bridge := network.NewLoopbackBridgeGroup()

	mix := flow.NewMix()
	f1 := flow.NewFlow("../captures/ping.cap")
	f2 := flow.NewFlow("../captures/youtube-short.cap")
	//TODO: Can only play one copy of each flow - need to re-write IPs so player
	// registrations don't collide
	mix.AddFlow(f1, 1)
	mix.AddFlow(f2, 1)

	player := flow.NewMixPlayer(mix, bridge)
	player.Play(15*time.Second)

	bridge.Shutdown(5*time.Second)
	bridgeStats := bridge.Stats()
	rxBytes := bridgeStats.Client.Rx.Bytes + bridgeStats.Server.Rx.Bytes
	txBytes := bridgeStats.Client.Tx.Bytes + bridgeStats.Server.Tx.Bytes
	totalBytes := rxBytes + txBytes
	fmt.Printf("Bytes  : rx: %v tx: %v\n", rxBytes, txBytes)
	fmt.Printf("Bitrate: %v\n", float64(totalBytes*8) / 15.0) 
}
