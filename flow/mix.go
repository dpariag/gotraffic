// A collection of flows that are to be replayed concurrently
package flow

import (
	"errors"
	"fmt"
)

type FlowGroup struct {
	Flow
	Copies uint64 // # of copies to be replayed concurrently
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
	fmt.Printf("Adding flow. %v packets. %v bytes %v ms. Avg packet size:%v\n",
		f.NumPackets(), f.NumBytes(), f.Duration(), f.NumBytes()/f.NumPackets())
	m.flows = append(m.flows, FlowGroup{*f, copies})
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

func (m *Mix) NextFlowGroup() (*FlowGroup, error) {
	if m.index < uint64(len(m.flows)) {
		fg := m.flows[m.index]
		m.index++
		return &fg, nil
	}
	m.index = 0
	return nil, errors.New("No more flow groups")
}
