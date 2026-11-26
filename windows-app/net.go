package main

// Talking to the Geode mod (port 47321) and to the Discord web extension (port 47322).

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// ---------- Geode mod ----------
type ModState struct {
	InLevel  bool    `json:"inLevel"`
	Percent  float64 `json:"percent"`
	Practice bool    `json:"practice"`
	Attempt  int64   `json:"attempt"`
}

var (
	modMu     sync.Mutex
	modState  ModState
	modLastOK time.Time
)

func pollMod() {
	client := &http.Client{
		Timeout:   300 * time.Millisecond,
		Transport: &http.Transport{Proxy: nil, DisableKeepAlives: true},
	}
	for {
		resp, err := client.Get("http://127.0.0.1:47321/state")
		if err == nil {
			var s ModState
			if json.NewDecoder(resp.Body).Decode(&s) == nil {
				modMu.Lock()
				modState, modLastOK = s, time.Now()
				modMu.Unlock()
			}
			resp.Body.Close()
			time.Sleep(10 * time.Millisecond)
		} else {
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func getMod() (ModState, bool) {
	modMu.Lock()
	defer modMu.Unlock()
	return modState, time.Since(modLastOK) < time.Second
}

// ---------- browser extension (discord.com deafen + instagram.com tab mute) ----------
var (
	webMu       sync.Mutex
	webSeq      int64
	discSeq     int64
	discAction  = "none" // deafen / undeafen
	instaSeq    int64
	instaAction = "none" // mute / unmute
	webWake     = make(chan struct{})
	webLastPoll time.Time
)

func wakeLocked() {
	webSeq++
	close(webWake)
	webWake = make(chan struct{})
}

func webDiscord(action string) {
	webMu.Lock()
	discSeq++
	discAction = action
	wakeLocked()
	webMu.Unlock()
}

func webInsta(action string) {
	webMu.Lock()
	instaSeq++
	instaAction = action
	wakeLocked()
	webMu.Unlock()
}

func webConnected() bool {
	webMu.Lock()
	defer webMu.Unlock()
	return time.Since(webLastPoll) < 25*time.Second
}

func serveWeb() error {
	ln, err := net.Listen("tcp", "127.0.0.1:47322")
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/wait", func(w http.ResponseWriter, r *http.Request) {
		known, _ := strconv.ParseInt(r.URL.Query().Get("seq"), 10, 64)
		webMu.Lock()
		webLastPoll = time.Now()
		wake, seq := webWake, webSeq
		webMu.Unlock()
		if known == seq {
			select {
			case <-wake:
			case <-time.After(20 * time.Second):
			case <-r.Context().Done():
			}
		}
		webMu.Lock()
		webLastPoll = time.Now()
		body := fmt.Sprintf(`{"seq":%d,"dseq":%d,"discord":%q,"iseq":%d,"insta":%q}`,
			webSeq, discSeq, discAction, instaSeq, instaAction)
		webMu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Cache-Control", "no-store")
		w.Write([]byte(body))
	})
	go http.Serve(ln, mux)
	return nil
}
