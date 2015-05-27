package stats

type Traffic struct {
	Bytes   uint64
	Packets uint64
}

type Directional struct {
	Rx Traffic
	Tx Traffic
}

type PlayerStats struct {
	Directional
	FlowsStarted uint64
	FlowsCompleted uint64
}

type BridgeGroupStats struct {
	Client Directional
	Server Directional
}

func (ps *PlayerStats) Add(other *PlayerStats) {
	ps.FlowsStarted += other.FlowsStarted
	ps.FlowsCompleted += other.FlowsCompleted
	ps.Rx.Packets += other.Rx.Packets
	ps.Tx.Packets += other.Tx.Packets
	ps.Rx.Bytes += other.Rx.Bytes
	ps.Tx.Bytes += other.Tx.Bytes
}
