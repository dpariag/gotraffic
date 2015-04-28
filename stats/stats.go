package stats

type Traffic struct {
	Bytes   uint64
	Packets uint64
}

type Directional struct {
	Rx Traffic
	Tx Traffic
}

type BridgeGroup struct {
	Client Directional
	Server Directional
}
