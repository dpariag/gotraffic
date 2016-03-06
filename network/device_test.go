package network

import (
	"testing"
	"time"
	"github.com/dpariag/gotraffic/flow"
)

func TestDeviceCreation(t *testing.T) {
	device := newDevice("test", clientDevice, newTestIOHandle())
	stats := device.Stats()

	if stats.Rx.Packets != 0 || stats.Rx.Bytes != 0 || stats.Tx.Packets != 0 || stats.Tx.Bytes != 0 {
		t.Errorf("Device stats are non-zero: %s\n", stats.String())
	}
}

func TestDeviceIO(t *testing.T) {
	io := newTestIOHandle() // simulates loopback
	device := newDevice("test", clientDevice, io)
	flow := flow.NewFlow("../captures/ping.cap")

	for _, pkt := range flow.Packets() {
		device.Send(pkt) // Sent pkts should also be received (loopback)
	}
	device.Shutdown(time.Second * 5)

	devStats := device.Stats()
	if devStats.Tx.Packets != flow.NumPackets() || devStats.Tx.Bytes != flow.NumBytes() {
		t.Errorf("Device sent %v packets/%v bytes. Should be %v packets/%v bytes\n",
			devStats.Tx.Packets, devStats.Tx.Bytes, flow.NumPackets(), flow.NumBytes())
	}

	if devStats.Rx.Packets != flow.NumPackets() || devStats.Rx.Bytes != flow.NumBytes() {
		t.Errorf("Device rcvd %v packets/%v bytes. Should be %v packets/%v bytes\n",
			devStats.Rx.Packets, devStats.Rx.Bytes, flow.NumPackets(), flow.NumBytes())
	}

	ioStats := io.Stats()
	if !devStats.Equals(&ioStats) {
		t.Errorf("Incorrect device stats.\n Device stats (%s)\n I/O handle stats (%s)\n",
			devStats.String(), ioStats.String())
	}
}
