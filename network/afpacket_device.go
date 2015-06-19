//+build linux

package network

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/afpacket"
	"time"
)

// Implements the ioHandle interface for packet I/O
type afpacketIOHandle struct {
	tpacket *afpacket.TPacket
}

func (h *afpacketIOHandle) ReadPacketData() (data []byte, ci gopacket.CaptureInfo, err error) {
	return h.tpacket.ReadPacketData()
}

func (h *afpacketIOHandle) WritePacketData(data []byte) (err error) {
	return h.tpacket.WritePacketData(data)
}

func newAfPacketIOHandle(name string) (handle *afpacketIOHandle, err error) {
	var af afpacketIOHandle
	af.tpacket, err = afpacket.NewTPacket(afpacket.OptInterface(name), afpacket.TPacketVersion2,
		afpacket.OptBlockTimeout(time.Millisecond*10000))
	return &af, err
}

func NewAfPacketDevice(name string, role int) Device {
	handle, err := newAfPacketIOHandle(name)
	if err != nil {
		panic(err)
	}
	d := newDevice(name, role, handle)
	return d
}
