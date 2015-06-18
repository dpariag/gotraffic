// +build darwin

package network

func NewBridgeGroup(iotype string, client string, server string) *bridgeGroup {
	if iotype != "pcap" {
		panic("Only PCAP packet I/O is supported on OSX")
	}
	return &bridgeGroup{client: NewPCAPDevice(client, clientDevice),
		server: NewPCAPDevice(server, serverDevice)}
}
