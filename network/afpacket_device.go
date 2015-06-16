package network

import (
	"github.com/google/gopacket/afpacket"
)

type afpacketDeviceHandle struct {
	tpacket	TPacket
}

func (h *afpacketDeviceHandle) ReadPacketData() (data []byte, ci gopacket.CaptureInfo, err error) {
	return h.tpacket.ReadPacketData()
}

func (h *afpacketDeviceHandle) WritePacketData(data []byte) (err error) {
	return h.tpacket.WritePacketData(data)
}
