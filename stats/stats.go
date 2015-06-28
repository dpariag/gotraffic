package stats

import (
	"fmt"
)

type trafficStats struct {
	Bytes   uint64
	Packets uint64
}

type Directional struct {
	Rx trafficStats
	Tx trafficStats
}

type FlowStats struct {
	Name string
	PlayerStats
}

type PlayerStats struct {
	Directional
	FlowsStarted   uint64
	FlowsCompleted uint64
}

type BridgeGroupStats struct {
	Client Directional
	Server Directional
}

type Stats interface {
	Stats() Directional
}

func (ps *PlayerStats) Add(other *PlayerStats) {
	ps.FlowsStarted += other.FlowsStarted
	ps.FlowsCompleted += other.FlowsCompleted
	ps.Rx.Packets += other.Rx.Packets
	ps.Tx.Packets += other.Tx.Packets
	ps.Rx.Bytes += other.Rx.Bytes
	ps.Tx.Bytes += other.Tx.Bytes
}

func (ps *PlayerStats) Clear() {
	ps.FlowsStarted = 0
	ps.FlowsCompleted = 0
	ps.Rx.Packets = 0
	ps.Tx.Packets = 0
	ps.Rx.Bytes = 0
	ps.Tx.Bytes = 0
}

func (d *Directional) Equals(other *Directional) bool {
	return d.Rx.Packets == other.Rx.Packets && d.Rx.Bytes == other.Rx.Bytes &&
		d.Tx.Packets == other.Tx.Packets && d.Tx.Bytes == other.Tx.Bytes
}

func (d *Directional) String() string {
	return fmt.Sprintf("Rx.Packets:%v Rx.Bytes:%v, Tx.Packets:%v Tx.Bytes%v",
		d.Rx.Packets, d.Rx.Bytes, d.Tx.Packets, d.Tx.Bytes)
}
