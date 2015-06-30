// A collection of flows that are to be replayed concurrently
package flow

type FlowGroup struct {
	*Flow
	Copies uint64 // # of copies of this flow in the mix
}

type Mix struct {
	flows    []FlowGroup
	numFlows uint64  // including all copies
	bitrate  float64 // bps of all concurrent flows
	index    uint64  // Index of next flow in the flow group
}

func NewMix() *Mix {
	return &Mix{}
}

func (m *Mix) AddFlow(f *Flow, copies uint64) {
	m.flows = append(m.flows, FlowGroup{f, copies})
	m.numFlows += copies
	m.bitrate += (f.Bitrate() * float64(copies))
}

func (m *Mix) Bitrate() float64 {
	return m.bitrate
}

func (m *Mix) NumFlows() uint64 {
	return m.numFlows
}

func (m *Mix) NumFlowGroups() uint64 {
	return uint64(len(m.flows))
}

func (m *Mix) FlowGroups() []FlowGroup {
	return m.flows
}
