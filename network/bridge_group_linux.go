// +build linux

package network

func NewBridgeGroup(iotype string, client string, server string) *bridgeGroup {
	if iotype == "pcap" {
		return &bridgeGroup{client: NewPCAPDevice(client, clientDevice),
			server: NewPCAPDevice(server, serverDevice)}
	} else if iotype == "afpacket" {
		return &bridgeGroup{client: NewAfPacketDevice(client, clientDevice),
			server: NewAfPacketDevice(server, serverDevice)}
	} else {
		panic("Unsupported I/O type " + iotype)
	}
}
