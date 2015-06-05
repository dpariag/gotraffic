package main

import (
	"os"
	"time"
	"fmt"
	"flag"
	"runtime"
	"net/http"
	"encoding/json"
	"git.svc.rocks/dpariag/gotraffic/flow"
	"git.svc.rocks/dpariag/gotraffic/stats"
	"git.svc.rocks/dpariag/gotraffic/player"
	"git.svc.rocks/dpariag/gotraffic/network"
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
	txBytes := stats.Tx.Bytes - lastStat.Tx.Bytes 
	elapsed := time.Since(lastTime).Seconds()
	txBitrate := float64(txBytes * 8) / elapsed 
	txBitrate = txBitrate / 1000000.0
	fmt.Fprintf(w, "Tx Bytes %v\n", txBytes)
	fmt.Fprintf(w, "Elapsed time %v\n", elapsed)
	fmt.Fprintf(w, "Tx bitrate %f\n", txBitrate)
	lastStat = stats
	lastTime = time.Now()
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
			fmt.Printf("Cur Tx bytes: %v\n", stats.Tx.Bytes)
			fmt.Printf("Last Tx bytes: %v\n", tx)
			fmt.Printf("Interval Tx bps: %v\n", (stats.Tx.Bytes - tx)*8)
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
	useLoopback := flag.Bool("use-loopback", false, "Use virtual loopback interfaces?")
	flag.Parse()

	var bridge network.BridgeGroup
	if *useLoopback {
		bridge = network.NewLoopbackBridgeGroup()
	} else if *subscriberDevice != "" && *internetDevice != "" {
		bridge = network.NewBridgeGroup(*subscriberDevice, *internetDevice)
	} else {
		fmt.Printf("No valid network interfaces specified\n")
		flag.Usage()
		os.Exit(1)
	}

	if *cores > 0 && *cores <= runtime.NumCPU() {
		runtime.GOMAXPROCS(*cores)
	}

	mix := flow.NewMix()
	mix.AddFlow(flow.NewFlow("../captures/youtube.cap"), 10)
	mix.AddFlow(flow.NewFlow("../captures/ping.cap"), 1)
	p = player.NewMixPlayer(mix, bridge)

	http.HandleFunc("/play", playHandler)
	http.HandleFunc("/stop", stopHandler)
	http.HandleFunc("/stats", statsHandler)
	http.Handle("/", http.FileServer(http.Dir(".")))
	http.Handle("/sse", handleSSE(p))
	http.ListenAndServe(":8080", nil)

	fmt.Printf("After ListenAndServe\n")
	bridge.Shutdown(1*time.Second)
}
