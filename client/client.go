package main

import "fmt"
import "time"
import "github.com/dpariag/flow"
import "github.com/dpariag/network"

func main() {
	iface := network.NewPCAPInterface("bridge0")
	iface.Init()

	mix := flow.NewMix()
	f1 := flow.NewFlow("./ping.cap")
	f2 := flow.NewFlow("./youtube-stream1.cap.pcap")
	// Should FlowGroup take the flow by value or pointer?
	mix.AddFlow(f1,1)
	mix.AddFlow(f2,1)

	player := flow.NewMixPlayer(mix, iface, 20*time.Second)
	player.Play()

	iface.Shutdown(5*time.Second)
	rxPkts, rxBytes := iface.RxStats()
	txPkts, txBytes := iface.TxStats()
	fmt.Printf("Packets: rx: %v tx: %v\n", rxPkts, txPkts)
	fmt.Printf("Bytes  : rx: %v tx: %v\n", rxBytes, txBytes)
}
