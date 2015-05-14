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

type BridgeGroup struct {
	Client Directional
	Server Directional
}
