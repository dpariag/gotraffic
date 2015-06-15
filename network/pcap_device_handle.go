package network

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

// Implements deviceHandle interface using pcap
type pcapDeviceHandle struct {
	pcap *pcap.Handle // Underlying PCAP device
}

func (p *pcapDeviceHandle) ReadPacketData() (data []byte, ci gopacket.CaptureInfo, err error) {
	return p.pcap.ReadPacketData()
}

func (p *pcapDeviceHandle) WritePacketData(data []byte) (err error) {
	return p.pcap.WritePacketData(data)
}

func newPCAPDeviceHandle(name string) (handle *pcapDeviceHandle, err error) {
	var p pcapDeviceHandle
	p.pcap, err = pcap.OpenLive(name, 2048, true, pcap.BlockForever)
	return &p, err
}

func NewPCAPDevice(name string, role int) Device {
	handle, err := newPCAPDeviceHandle(name)
	if err != nil {
		panic(err)
	}
	d := newDevice(name, role, handle)
	return d
}
