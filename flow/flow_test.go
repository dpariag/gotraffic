package flow

import (
	"testing"
	"time"
)

type flowInfo struct {
	numPkts		uint64
	duration	time.Duration
	bytes		uint64
}

func verifyFlow(t *testing.T, pcapPath string, info *flowInfo) {
	t.Logf("Test flow: %v\n", pcapPath)
	f := NewFlow(pcapPath)

	if uint64(len(f.pkts)) != info.numPkts {
		t.Errorf("Incorrect pkt count. Expected: %v. Found %v\n",info.numPkts, len(f.pkts))
	}
	if f.duration != info.duration {
		t.Errorf("Incorrect flow duration. Expected: %v. Found: %v\n", info.duration, f.duration)
	}
	if f.NumBytes() != info.bytes {
		t.Errorf("Incorrect byte count. Expected: %v. Found: %v\n", info.bytes, f.NumBytes())
	}
	expectedBitrate := float64(info.bytes * 8) / float64(info.duration.Seconds())
	if f.Bitrate() != expectedBitrate {
		t.Errorf("Incorrect bitrate. Expected: %v. Found: %v\n", expectedBitrate, f.Bitrate())
	}
}

func TestICMPFlow(t *testing.T) {
	verifyFlow(t, "captures/ping.cap", &flowInfo{14, 6027973000, 1372})
}

func TestYouTubeFlow(t *testing.T) {
	verifyFlow(t, "captures/youtube.cap", &flowInfo{87, 4829350000, 7560})
}
