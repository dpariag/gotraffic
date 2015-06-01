package main

import (
	"time"
	"fmt"
	"net/http"
	"encoding/json"
	"git.svc.rocks/dpariag/gotraffic/flow"
	"git.svc.rocks/dpariag/gotraffic/player"
	"git.svc.rocks/dpariag/gotraffic/network"
)

var p *player.MixPlayer

func playHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Starting traffic")
	p.Play()
}

func stopHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Stopping traffic")
	p.Stop()
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	stats := p.Stats()
	fmt.Fprintf(w, "Started %v flows\n", stats.FlowsStarted)
	fmt.Fprintf(w, "Completed %v flows\n", stats.FlowsCompleted)
	fmt.Fprintf(w, "Sent %v packets\n", stats.Tx.Packets)
	fmt.Fprintf(w, "Rcvd %v packets\n", stats.Rx.Packets)
	fmt.Fprintf(w, "Sent %v bytes\n", stats.Tx.Bytes)
	fmt.Fprintf(w, "Rcvd %v bytes\n", stats.Rx.Bytes)
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
		for {
			stats := player.Stats()
			ts := time.Now()
			tx := stats.Tx.Bytes
			rx := stats.Rx.Bytes
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
	bridge := network.NewLoopbackBridgeGroup()
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
