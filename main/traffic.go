package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"git.svc.rocks/dpariag/gotraffic/flow"
	"git.svc.rocks/dpariag/gotraffic/network"
	"git.svc.rocks/dpariag/gotraffic/player"
	"git.svc.rocks/dpariag/gotraffic/stats"
	"net/http"
	"os"
	"runtime"
	"time"
)

var p *player.MixPlayer
var lastStat stats.PlayerStats
var lastTime time.Time

func playHandler(w http.ResponseWriter, r *http.Request) {
	p.Play()
	lastTime = time.Now()
}

func stopHandler(w http.ResponseWriter, r *http.Request) {
	p.Stop()
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	stats := p.Stats()
	json, _ := json.Marshal(stats)
	fmt.Fprintf(w, "data: %s\n\n", json)
}

func flowStatsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Flow stats handler")
	flowStats := p.FlowStats()
	json,_ := json.Marshal(flowStats)
	fmt.Fprintf(w, "data: %s\n\n", json)
}

func handleSSE(player *player.MixPlayer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "SSE not available", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		var tx uint64
		for {
			stats := player.Stats()
			fmt.Printf("Interval Tx (Mbps): %v\n", ((stats.Tx.Bytes-tx)*8)/1000000.0)
			ts := time.Now()
			tx = stats.Tx.Bytes * 8
			rx := stats.Rx.Bytes * 8
			ev := map[string]interface{}{
				"time": int(ts.Unix()),
				"tx":   tx,
				"rx":   rx,
			}
			b, err := json.Marshal(&ev)
			if err != nil {
				panic("can't encode?")
			}
			_, err = fmt.Fprintf(w, "data: %s\n\n", b)
			if err != nil {
				break // Client disconnected.
			}
			conn.Flush()
			time.Sleep(time.Second)
		}
	}
}

func main() {
	cores := flag.Int("num-cores", 2, "Number of logical CPUs available for use")
	subscriberDevice := flag.String("sub-interface", "", "Subscriber interface")
	internetDevice := flag.String("ext-interface", "", "External interface")
	ioType := flag.String("io-type", "pcap", "Packet I/O mechanism: pcap, afpacket")
	useLoopback := flag.Bool("use-loopback", false, "Use virtual loopback interfaces?")
	flag.Parse()

	var bridge network.BridgeGroup
	if *useLoopback {
		bridge = network.NewLoopbackBridgeGroup()
	} else if *subscriberDevice != "" && *internetDevice != "" {
		bridge = network.NewBridgeGroup(*ioType, *subscriberDevice, *internetDevice)
	} else {
		fmt.Printf("No valid network interfaces specified\n")
		flag.Usage()
		os.Exit(1)
	}

	if *cores > 0 && *cores <= runtime.NumCPU() {
		runtime.GOMAXPROCS(*cores)
	}

	mix := flow.NewMix()
	mix.AddFlow(flow.NewFlow("../captures/youtube.cap"), 20)
	mix.AddFlow(flow.NewFlow("../captures/ping.cap"), 1)
	p = player.NewMixPlayer(mix, bridge)

	http.HandleFunc("/play", playHandler)
	http.HandleFunc("/stop", stopHandler)
	http.HandleFunc("/stats", statsHandler)
	http.HandleFunc("/stats/flows", flowStatsHandler)
	http.Handle("/", http.FileServer(http.Dir(".")))
	http.Handle("/sse", handleSSE(p))
	http.ListenAndServe(":80", nil)
	bridge.Shutdown(1 * time.Second)
}
