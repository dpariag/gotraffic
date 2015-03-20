package main

import "os"
import "fmt"
import "code.google.com/p/gopacket"
import "code.google.com/p/gopacket/pcap"

func main() {
	iface := os.Args[1]
	fmt.Println("Starting server...\n")

	ifaceHandle,err := pcap.OpenLive(iface, 2048, true, 60)
	if err != nil {
		panic(err)
	}

	fmt.Println("Listening for packets.")
	packetSource := gopacket.NewPacketSource(ifaceHandle, ifaceHandle.LinkType())
	for packet:= range packetSource.Packets() {
		fmt.Printf("%v\n", packet.String())
	}
}
