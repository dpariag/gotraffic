package flow

import (
	"testing"
)

func TestSingleFlowMix(t *testing.T) {
	m := NewMix()
	f := NewFlow("../captures/ping.cap")
	m.AddFlow(f, 1)

	if m.NumFlows() != 1 {
		t.Errorf("Incorrect number of flows. Expected %v, Found: %v\n", 1, m.NumFlows())
	}

	if m.Bitrate() != f.Bitrate() {
		t.Errorf("Incorrect Mix bitrate. Expected: %v, Found: %v\n", f.Bitrate(), m.Bitrate())
	}
}
